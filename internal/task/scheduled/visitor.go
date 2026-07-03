// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package scheduled

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-co-op/gocron/v2"
	visitorModel "github.com/lin-snow/ech0/internal/model/visitor"
	visitorRepository "github.com/lin-snow/ech0/internal/repository/visitor"
	"github.com/lin-snow/ech0/internal/visitor"
	logUtil "github.com/lin-snow/ech0/pkg/log"
)

const visitorSnapshotTag = "VisitorSnapshotSchedule"

// VisitorSnapshot 每 60 分钟把内存中的当日 PV/UV upsert 落库，并清理超出 7 天窗口的历史行。
// 实现 task.StopHook：优雅退出由 OnStop 补一次快照，非优雅退出最多丢 1 小时数据。
type VisitorSnapshot struct {
	tracker *visitor.Tracker
	repo    *visitorRepository.VisitorRepository
}

func NewVisitorSnapshot(tracker *visitor.Tracker, repo *visitorRepository.VisitorRepository) *VisitorSnapshot {
	return &VisitorSnapshot{tracker: tracker, repo: repo}
}

func (v *VisitorSnapshot) Name() string { return "visitor-snapshot" }

// Schedule 先把最近 7 天历史灌回 tracker，再挂上每 60 分钟的快照 + 清理作业。
func (v *VisitorSnapshot) Schedule(ctx context.Context, s gocron.Scheduler) error {
	if err := v.restore(ctx); err != nil {
		return err
	}

	_, err := s.NewJob(
		gocron.DurationJob(60*time.Minute),
		gocron.NewTask(func() {
			v.flush(context.Background())
			cutoff := cutoffDate(time.Now().UTC())
			if err := v.repo.DeleteOlderThan(context.Background(), cutoff); err != nil {
				logUtil.GetLogger().Error("Failed to cleanup visitor stats",
					slog.String("module", logModule), logUtil.Err(err))
			}
		}),
		gocron.WithTags(visitorSnapshotTag),
	)
	if err != nil {
		logUtil.GetLogger().Error("Failed to schedule visitor snapshot task",
			slog.String("module", logModule), logUtil.Err(err))
	}
	return err
}

// OnStop 优雅退出时补一次快照。用 background ctx，避免停机 ctx 已取消导致最后一次落盘失败。
func (v *VisitorSnapshot) OnStop(context.Context) {
	v.flush(context.Background())
}

// flush 将当前内存中的今日 PV/UV upsert 到数据库。
func (v *VisitorSnapshot) flush(ctx context.Context) {
	if v.tracker == nil || v.repo == nil {
		return
	}
	today := v.tracker.TodayStat()
	if err := v.repo.UpsertDailyStat(ctx, buildDailyStat(today)); err != nil {
		logUtil.GetLogger().Error("Failed to upsert visitor daily stat",
			slog.String("module", logModule), logUtil.Err(err))
	}
}

func (v *VisitorSnapshot) restore(ctx context.Context) error {
	if v.tracker == nil || v.repo == nil {
		return nil
	}
	stats, err := v.repo.GetRecentDays(ctx, 7)
	if err != nil {
		logUtil.GetLogger().Error("Failed to load visitor stats",
			slog.String("module", logModule), logUtil.Err(err))
		return err
	}
	if len(stats) == 0 {
		return nil
	}
	v.tracker.LoadHistory(convertHistory(stats))
	return nil
}

// buildDailyStat 把内存态 DayStat 映射为可落库的 DailyStat。
func buildDailyStat(stat visitor.DayStat) visitorModel.DailyStat {
	return visitorModel.DailyStat{Date: stat.Date, PV: stat.PV, UV: stat.UV}
}

// convertHistory 把落库历史映射回内存态，供 tracker 启动期回灌。
func convertHistory(stats []visitorModel.DailyStat) []visitor.DayStat {
	history := make([]visitor.DayStat, 0, len(stats))
	for _, s := range stats {
		history = append(history, visitor.DayStat{Date: s.Date, PV: s.PV, UV: s.UV})
	}
	return history
}

// cutoffDate 与 visitor.Tracker 的 keepDays(7) 对齐：保留今天及前 6 天，共 7 天。
func cutoffDate(now time.Time) string {
	return now.UTC().AddDate(0, 0, -6).Format("2006-01-02")
}
