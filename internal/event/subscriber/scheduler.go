package subscriber

import (
	"context"

	contracts "github.com/lin-snow/ech0/internal/event/contracts"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
)

type BackupScheduleApplier interface {
	ApplyBackupSchedule(schedule settingModel.BackupSchedule) error
}

type BackupScheduler struct {
	applier BackupScheduleApplier
}

func NewBackupScheduler(applier BackupScheduleApplier) *BackupScheduler {
	return &BackupScheduler{applier: applier}
}

func (bs *BackupScheduler) HandleBackupScheduleUpdated(
	ctx context.Context,
	e contracts.UpdateBackupScheduleEvent,
) error {
	_ = ctx
	return bs.applier.ApplyBackupSchedule(e.Schedule)
}
