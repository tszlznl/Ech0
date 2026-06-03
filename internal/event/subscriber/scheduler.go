// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package subscriber

import (
	"context"

	"github.com/lin-snow/ech0/internal/event"
	eventbus "github.com/lin-snow/ech0/internal/event/bus"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
)

type SnapshotScheduleApplier interface {
	ApplySnapshotSchedule(schedule settingModel.SnapshotSchedule) error
}

type SnapshotScheduler struct {
	applier SnapshotScheduleApplier
}

func NewSnapshotScheduler(applier SnapshotScheduleApplier) *SnapshotScheduler {
	return &SnapshotScheduler{applier: applier}
}

func (ss *SnapshotScheduler) HandleSnapshotScheduleUpdated(
	ctx context.Context,
	e event.UpdateSnapshotSchedule,
) error {
	_ = ctx
	return ss.applier.ApplySnapshotSchedule(e.Schedule)
}

func (ss *SnapshotScheduler) Registrations() []eventbus.Registration {
	return []eventbus.Registration{
		eventbus.On(ss.HandleSnapshotScheduleUpdated, eventbus.AsyncSequential()...),
	}
}
