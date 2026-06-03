// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package scheduled

import (
	"context"
	"strings"
	"sync"

	"github.com/go-co-op/gocron/v2"
	contracts "github.com/lin-snow/ech0/internal/event/contracts"
	publisher "github.com/lin-snow/ech0/internal/event/publisher"
	coreMigrator "github.com/lin-snow/ech0/internal/migrator"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	settingService "github.com/lin-snow/ech0/internal/service/setting"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"go.uber.org/zap"
)

const snapshotScheduleTag = "SnapshotSchedule"

// Snapshot 定时创建系统快照（产出统一 Snapshot，含尽力 S3 上传）。支持运行期通过
// ApplySnapshotSchedule 动态重配，实现 eventsubscriber.SnapshotScheduleApplier（由 task.Find 取出
// 供 SnapshotScheduler 订阅者使用）。打包 + S3 的执行逻辑统一收敛到 migrator.ExportEngine，定时
// 快照不走 job.Manager（无需 UI 状态/取消，且避免与手动导出抢占同一作业行）。
type Snapshot struct {
	settingService settingService.Service
	exporter       *coreMigrator.ExportEngine
	publisher      *publisher.Publisher

	// mu 同时保护 scheduler 字段与「移除旧作业 + 挂新作业」的重配过程。
	mu        sync.Mutex
	scheduler gocron.Scheduler // Schedule 时捕获，供运行期 ApplySnapshotSchedule 使用
}

func NewSnapshot(
	settingSvc settingService.Service,
	exporter *coreMigrator.ExportEngine,
	publisher *publisher.Publisher,
) *Snapshot {
	return &Snapshot{settingService: settingSvc, exporter: exporter, publisher: publisher}
}

func (b *Snapshot) Name() string { return "snapshot" }

// Schedule 捕获 scheduler，读取计划设置；启用时按 cron 挂上定时快照作业。
func (b *Snapshot) Schedule(_ context.Context, s gocron.Scheduler) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.scheduler = s

	var setting settingModel.SnapshotSchedule
	if err := b.settingService.GetSnapshotScheduleSetting(&setting); err != nil {
		logUtil.GetLogger().Error("Failed to get snapshot schedule setting",
			zap.String("module", logModule), zap.Error(err))
		// 读取失败时默认关闭定时快照任务
		setting.Enable = false
		setting.CronExpression = "0 2 * * 0" // 每周日2点执行一次
	}
	if !setting.Enable {
		return nil
	}
	return b.scheduleJob(setting.CronExpression)
}

// ApplySnapshotSchedule 在运行期动态更新定时快照任务（实现 eventsubscriber.SnapshotScheduleApplier）。
func (b *Snapshot) ApplySnapshotSchedule(schedule settingModel.SnapshotSchedule) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.scheduler == nil {
		// 尚未 Schedule（Manager 未 Start），无需处理。
		return nil
	}

	logUtil.GetLogger().Info("Applying snapshot schedule",
		zap.String("module", logModule),
		zap.Bool("enable", schedule.Enable),
		zap.String("cron", schedule.CronExpression),
	)

	// 先移除旧任务，避免重复触发。
	b.scheduler.RemoveByTags(snapshotScheduleTag)
	if !schedule.Enable {
		logUtil.GetLogger().Info("Snapshot schedule disabled, jobs removed", zap.String("module", logModule))
		return nil
	}

	if err := b.scheduleJob(schedule.CronExpression); err != nil {
		logUtil.GetLogger().Error("Failed to apply snapshot schedule",
			zap.String("module", logModule), zap.Error(err))
		return err
	}
	logUtil.GetLogger().Info("Snapshot schedule applied successfully", zap.String("module", logModule))
	return nil
}

// scheduleJob 按 cron 表达式挂上一个带 tag 的定时快照作业。调用方须持有 b.mu。
func (b *Snapshot) scheduleJob(cronExpression string) error {
	// 判断 cron 表达式的字段数量来确定是否包含秒字段：
	// 5 位（分 时 日 月 周）withSeconds=false；6 位（秒 分 时 日 月 周）withSeconds=true。
	withSeconds := len(strings.Fields(cronExpression)) == 6

	_, err := b.scheduler.NewJob(
		gocron.CronJob(cronExpression, withSeconds),
		gocron.NewTask(func() {
			ctx := context.Background()

			if _, err := b.exporter.Export(ctx, func(string, any) {}); err != nil {
				logUtil.GetLogger().Error("Failed to execute scheduled snapshot",
					zap.String("module", logModule),
					zap.Error(err))
				return
			}

			if err := b.publisher.SystemSnapshot(
				ctx,
				contracts.SystemSnapshotEvent{Info: "System scheduled snapshot completed"},
			); err != nil {
				logUtil.GetLogger().Error("Failed to publish snapshot completed event",
					zap.String("module", logModule), zap.Error(err))
			}
		}),
		gocron.WithTags(snapshotScheduleTag),
	)
	if err != nil {
		logUtil.GetLogger().Error("Failed to schedule snapshot task",
			zap.String("module", logModule), zap.Error(err))
	}
	return err
}
