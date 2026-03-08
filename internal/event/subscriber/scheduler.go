package subscriber

import (
	"context"

	contracts "github.com/lin-snow/ech0/internal/event/contracts"
)

type BackupScheduler struct{}

func NewBackupScheduler() *BackupScheduler {
	return &BackupScheduler{}
}

func (bs *BackupScheduler) HandleBackupScheduleUpdated(
	ctx context.Context,
	e contracts.UpdateBackupScheduleEvent,
) error {
	_ = ctx
	_ = e
	return nil
}
