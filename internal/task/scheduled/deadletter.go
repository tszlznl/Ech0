// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package scheduled

import (
	"context"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/lin-snow/ech0/internal/event"
	eventbus "github.com/lin-snow/ech0/internal/event/bus"
	queueRepository "github.com/lin-snow/ech0/internal/repository/queue"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"github.com/lin-snow/ech0/pkg/busen"
	"go.uber.org/zap"
)

// DeadLetter 周期消费死信队列，重新发布事件以触发重试。
type DeadLetter struct {
	queueRepo *queueRepository.QueueRepository
	bus       *busen.Bus
}

func NewDeadLetter(queueRepo *queueRepository.QueueRepository, busProvider func() *busen.Bus) *DeadLetter {
	return &DeadLetter{queueRepo: queueRepo, bus: busProvider()}
}

func (d *DeadLetter) Name() string { return "dead-letter-consume" }

// Schedule 每 5 分钟取出死信逐个重试，保证死信重试能在分钟级恢复。
func (d *DeadLetter) Schedule(_ context.Context, s gocron.Scheduler) error {
	_, err := s.NewJob(
		gocron.DurationJob(5*time.Minute),
		gocron.NewTask(func() {
			// 取出死信队列中的任务，逐个重试
			deadLetters, err := d.queueRepo.ListDeadLetters(context.Background(), 10)
			if err != nil {
				logUtil.GetLogger().Error("Failed To Get DeadLetters!",
					zap.String("module", logModule), zap.Error(err))
			}

			// 遍历死信任务，重新发布事件触发重试
			for _, dl := range deadLetters {
				eventbus.Notify(context.Background(), d.bus, event.DeadLetterRetried{DeadLetter: dl})
			}
		}),
	)
	if err != nil {
		logUtil.GetLogger().Error("Failed to schedule dead letter consume task",
			zap.String("module", logModule), zap.Error(err))
	}
	return err
}
