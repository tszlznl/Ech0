// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package scheduled

import (
	"context"
	"strings"
	"sync"

	"github.com/go-co-op/gocron/v2"
	"github.com/lin-snow/ech0/internal/backup"
	contracts "github.com/lin-snow/ech0/internal/event/contracts"
	publisher "github.com/lin-snow/ech0/internal/event/publisher"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	settingService "github.com/lin-snow/ech0/internal/service/setting"
	"github.com/lin-snow/ech0/internal/storage"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"go.uber.org/zap"
)

const backupScheduleTag = "BackupSchedule"

// Backup 定时执行系统备份并可选上传 S3。支持运行期通过 ApplyBackupSchedule 动态重配，
// 实现 eventsubscriber.BackupScheduleApplier（由 task.Find 取出供 BackupScheduler 订阅者使用）。
type Backup struct {
	settingService settingService.Service
	storageManager *storage.Manager
	publisher      *publisher.Publisher

	// mu 同时保护 scheduler 字段与「移除旧作业 + 挂新作业」的重配过程。
	mu        sync.Mutex
	scheduler gocron.Scheduler // Schedule 时捕获，供运行期 ApplyBackupSchedule 使用
}

func NewBackup(
	settingSvc settingService.Service,
	storageManager *storage.Manager,
	publisher *publisher.Publisher,
) *Backup {
	return &Backup{settingService: settingSvc, storageManager: storageManager, publisher: publisher}
}

func (b *Backup) Name() string { return "backup" }

// Schedule 捕获 scheduler，读取备份设置；启用时按 cron 挂上备份作业。
func (b *Backup) Schedule(_ context.Context, s gocron.Scheduler) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.scheduler = s

	var setting settingModel.BackupSchedule
	if err := b.settingService.GetBackupScheduleSetting(&setting); err != nil {
		logUtil.GetLogger().Error("Failed to get backup schedule setting",
			zap.String("module", logModule), zap.Error(err))
		// 读取失败时默认关闭定时备份任务
		setting.Enable = false
		setting.CronExpression = "0 2 * * 0" // 每周日2点执行一次
	}
	if !setting.Enable {
		return nil
	}
	return b.scheduleJob(setting.CronExpression)
}

// ApplyBackupSchedule 在运行期动态更新备份任务（实现 eventsubscriber.BackupScheduleApplier）。
func (b *Backup) ApplyBackupSchedule(schedule settingModel.BackupSchedule) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.scheduler == nil {
		// 尚未 Schedule（Manager 未 Start），无需处理。
		return nil
	}

	logUtil.GetLogger().Info("Applying backup schedule",
		zap.String("module", logModule),
		zap.Bool("enable", schedule.Enable),
		zap.String("cron", schedule.CronExpression),
	)

	// 先移除旧任务，避免重复触发。
	b.scheduler.RemoveByTags(backupScheduleTag)
	if !schedule.Enable {
		logUtil.GetLogger().Info("Backup schedule disabled, jobs removed", zap.String("module", logModule))
		return nil
	}

	if err := b.scheduleJob(schedule.CronExpression); err != nil {
		logUtil.GetLogger().Error("Failed to apply backup schedule",
			zap.String("module", logModule), zap.Error(err))
		return err
	}
	logUtil.GetLogger().Info("Backup schedule applied successfully", zap.String("module", logModule))
	return nil
}

// scheduleJob 按 cron 表达式挂上一个带 tag 的备份作业。调用方须持有 b.mu。
func (b *Backup) scheduleJob(cronExpression string) error {
	// 判断 cron 表达式的字段数量来确定是否包含秒字段：
	// 5 位（分 时 日 月 周）withSeconds=false；6 位（秒 分 时 日 月 周）withSeconds=true。
	withSeconds := len(strings.Fields(cronExpression)) == 6

	_, err := b.scheduler.NewJob(
		gocron.CronJob(cronExpression, withSeconds),
		gocron.NewTask(func() {
			ctx := context.Background()

			backupPath, fileName, err := backup.ExecuteBackup()
			if err != nil {
				logUtil.GetLogger().Error("Failed to execute scheduled backup",
					zap.String("module", logModule),
					zap.String("path", backupPath),
					zap.String("fileName", fileName),
					zap.Error(err))
				return
			}

			b.tryUploadToS3(ctx, backupPath, fileName)

			if err := b.publisher.SystemBackup(
				ctx,
				contracts.SystemBackupEvent{Info: "System scheduled backup completed"},
			); err != nil {
				logUtil.GetLogger().Error("Failed to publish backup completed event",
					zap.String("module", logModule), zap.Error(err))
			}
		}),
		gocron.WithTags(backupScheduleTag),
	)
	if err != nil {
		logUtil.GetLogger().Error("Failed to schedule backup task",
			zap.String("module", logModule), zap.Error(err))
	}
	return err
}

func (b *Backup) tryUploadToS3(ctx context.Context, backupPath, fileName string) {
	if b.storageManager == nil {
		return
	}
	selector := b.storageManager.GetSelector()
	if selector == nil || !selector.ObjectEnabled() {
		return
	}

	cfg := b.storageManager.GetStorageConfig(ctx)
	if err := backup.UploadToS3(ctx, backupPath, fileName, cfg); err != nil {
		logUtil.GetLogger().Warn("Failed to upload backup to S3",
			zap.String("module", logModule), zap.Error(err))
	}
}
