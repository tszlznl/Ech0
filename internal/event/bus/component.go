// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package bus

import (
	"context"
	"errors"

	busen "github.com/lin-snow/Busen"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"go.uber.org/zap"
)

type EventBus struct {
	bus *busen.Bus
}

func NewEventBus(busProvider func() *busen.Bus) *EventBus {
	return &EventBus{bus: busProvider()}
}

func (c *EventBus) Name() string {
	return "event_bus"
}

func (c *EventBus) Start(context.Context) error {
	return nil
}

func (c *EventBus) Stop(ctx context.Context) error {
	if c.bus == nil {
		return errors.New("event bus is nil")
	}
	result, err := c.bus.Shutdown(ctx, busen.ShutdownDrain)
	if err != nil {
		return err
	}
	logUtil.GetLogger().Info("event bus shutdown",
		zap.Bool("completed", result.Completed),
		zap.Int64("processed", result.Processed),
		zap.Int64("dropped", result.Dropped),
		zap.Int64("rejected", result.Rejected))
	return nil
}
