// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package subscriber

import (
	"context"

	contracts "github.com/lin-snow/ech0/internal/event/contracts"
	registry "github.com/lin-snow/ech0/internal/event/registry"
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

func (bs *SnapshotScheduler) HandleSnapshotScheduleUpdated(
	ctx context.Context,
	e contracts.UpdateSnapshotScheduleEvent,
) error {
	_ = ctx
	return bs.applier.ApplySnapshotSchedule(e.Schedule)
}

func (bs *SnapshotScheduler) Subscriptions() []registry.Subscription {
	return []registry.Subscription{
		registry.TopicSubscription(
			contracts.TopicSnapshotScheduleUpdate,
			bs.HandleSnapshotScheduleUpdated,
			registry.SystemSubscribeOptions()...,
		),
	}
}
