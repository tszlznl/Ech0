// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package scheduled

import (
	"context"
	"strings"
	"sync"

	"github.com/go-co-op/gocron/v2"
	"github.com/lin-snow/ech0/internal/event"
	eventbus "github.com/lin-snow/ech0/internal/event/bus"
	"github.com/lin-snow/ech0/internal/kvstore"
	coreMigrator "github.com/lin-snow/ech0/internal/migrator"
	coreSetting "github.com/lin-snow/ech0/internal/setting"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"github.com/lin-snow/ech0/pkg/busen"
	"go.uber.org/zap"
)

const snapshotScheduleTag = "SnapshotSchedule"

// Snapshot 定时创建系统快照（产出统一 Snapshot，含尽力 S3 上传）。它自管「调度 + 订阅」整个
// 生命周期：Schedule 时捕获 scheduler 并订阅 UpdateSnapshotSchedule，收到即 reload；OnStop 退订。
// 计划配置统一经 setting 引擎读 durableKV（而非依赖整个 SettingService），从根上断开
// 「SettingService → Snapshot → SettingService」的构造环，也无需跨注入器的订阅者壳 / 反射查找。
// 打包 + S3 的执行收敛到 migrator.ExportEngine，定时快照不走 job.Manager（无需 UI 状态/取消，
// 且避免与手动导出抢占同一作业行）。
type Snapshot struct {
	durableKV kvstore.Store
	exporter  *coreMigrator.ExportEngine
	bus       *busen.Bus

	// mu 同时保护 scheduler/unsub 字段与「移除旧作业 + 挂新作业」的重配过程。
	mu        sync.Mutex
	scheduler gocron.Scheduler // Schedule 时捕获，供 reload 使用
	unsub     func()           // 总线订阅的退订句柄，OnStop 时调用
}

func NewSnapshot(
	durableKV kvstore.Store,
	exporter *coreMigrator.ExportEngine,
	busProvider func() *busen.Bus,
) *Snapshot {
	return &Snapshot{durableKV: durableKV, exporter: exporter, bus: busProvider()}
}

func (b *Snapshot) Name() string { return "snapshot" }

// Schedule 捕获 scheduler，订阅运行期计划变更，并按当前计划挂上定时快照作业。
func (b *Snapshot) Schedule(ctx context.Context, s gocron.Scheduler) error {
	b.mu.Lock()
	b.scheduler = s
	b.mu.Unlock()

	if err := b.subscribe(); err != nil {
		return err
	}
	return b.reload(ctx)
}

// OnStop 退订总线，避免停机后残留订阅。实现 task.StopHook。
func (b *Snapshot) OnStop(_ context.Context) {
	b.mu.Lock()
	unsub := b.unsub
	b.unsub = nil
	b.mu.Unlock()
	if unsub != nil {
		unsub()
	}
}

// subscribe 把自己挂上总线：收到 UpdateSnapshotSchedule 即按持久化的最新计划重配（保序消费）。
func (b *Snapshot) subscribe() error {
	unsub, err := eventbus.On(b.handleScheduleChanged, eventbus.AsyncSequential()...)(b.bus)
	if err != nil {
		return err
	}
	b.mu.Lock()
	b.unsub = unsub
	b.mu.Unlock()
	return nil
}

// handleScheduleChanged 忽略事件载荷，直接重读持久化计划——以「存了什么」为唯一真相源，天然幂等、
// 对并发更新收敛。事件载荷仅供 webhook 桥接使用。
func (b *Snapshot) handleScheduleChanged(ctx context.Context, _ event.UpdateSnapshotSchedule) error {
	return b.reload(ctx)
}

// reload 是唯一的「应用计划」路径：读当前计划 → 移除旧作业 → 启用则按 cron 重新挂上。
// 启动（Schedule）与运行期（事件）都走它，故 withSeconds 解析、enable 判断、tag 只有一份。
// 读取失败时保留现有作业（不贸然移除），仅记录并上抛错误，避免一次瞬时读故障误关定时快照。
func (b *Snapshot) reload(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.scheduler == nil {
		// 尚未 Schedule（Manager 未 Start），无需处理。
		return nil
	}

	schedule, err := coreSetting.Get(ctx, b.durableKV, coreSetting.Snapshot)
	if err != nil {
		logUtil.GetLogger().Error("Failed to read snapshot schedule, keeping current jobs",
			zap.String("module", logModule), zap.Error(err))
		return err
	}

	// 先移除旧作业，避免重复触发（启动时无旧作业，是无害 no-op）。
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
	logUtil.GetLogger().Info("Snapshot schedule applied",
		zap.String("module", logModule), zap.String("cron", schedule.CronExpression))
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

			eventbus.Notify(ctx, b.bus, event.SystemSnapshot{Info: "System scheduled snapshot completed"})
		}),
		gocron.WithTags(snapshotScheduleTag),
		// 单例防重叠：上一次快照还在跑时，本次触发直接跳过而非排队，避免大数据/慢 S3 下
		// 快照叠跑或堆积成 backlog（Reschedule = 丢弃本 tick，等下一个 cron 点）。
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	)
	if err != nil {
		logUtil.GetLogger().Error("Failed to schedule snapshot task",
			zap.String("module", logModule), zap.Error(err))
	}
	return err
}
