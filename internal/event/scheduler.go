package event

import (
	"context"
)

type BackupScheduler struct{}

func NewBackupScheduler() *BackupScheduler {
	return &BackupScheduler{}
}

func (bs *BackupScheduler) HandleBackupScheduleUpdated(
	ctx context.Context,
	e UpdateBackupScheduleEvent,
) error {
	_ = ctx
	_ = e
	// TODO: 这里可进一步接入 Tasker 的动态重载能力。
	return nil
}
