// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package bus

import (
	"context"

	"github.com/lin-snow/ech0/internal/event"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"github.com/lin-snow/ech0/pkg/busen"
	"go.uber.org/zap"
)

// registerWebhookObservers 为所有可观测事件注册同步订阅：把强类型事件转为中立 WebhookObservation
// 并投递给 observer。并发 / 重试由 webhook Dispatcher 的 worker pool 承担，故此处同步即可。
func registerWebhookObservers(
	b *busen.Bus,
	deliver func(ctx context.Context, obs event.WebhookObservation) error,
) ([]func(), error) {
	registrations := []Registration{
		observeWebhook[event.UserCreated](deliver),
		observeWebhook[event.UserUpdated](deliver),
		observeWebhook[event.UserDeleted](deliver),
		observeWebhook[event.EchoCreated](deliver),
		observeWebhook[event.EchoUpdated](deliver),
		observeWebhook[event.EchoDeleted](deliver),
		observeWebhook[event.CommentCreated](deliver),
		observeWebhook[event.CommentStatusUpdated](deliver),
		observeWebhook[event.CommentDeleted](deliver),
		observeWebhook[event.ResourceUploaded](deliver),
		observeWebhook[event.SystemSnapshot](deliver),
		observeWebhook[event.SystemExport](deliver),
		observeWebhook[event.UpdateSnapshotSchedule](deliver),
	}

	var unsubs []func()
	for _, reg := range registrations {
		unsub, err := reg(b)
		if err != nil {
			for _, u := range unsubs {
				u()
			}
			return nil, err
		}
		unsubs = append(unsubs, unsub)
	}
	return unsubs, nil
}

// observeWebhook 构造单个事件类型的 webhook 观察订阅（同步）。它需要 busen.Event 的 Meta（即
// source=ech0 等元数据），故不复用 On —— On 只透传 Value。
func observeWebhook[T event.Named](
	deliver func(ctx context.Context, obs event.WebhookObservation) error,
) Registration {
	return func(b *busen.Bus) (func(), error) {
		return busen.Subscribe(b, func(ctx context.Context, e busen.Event[T]) error {
			obs, err := event.NewWebhookObservation(e.Value.EventName(), e.Value, e.Meta)
			if err != nil {
				logUtil.GetLogger().Warn("build webhook observation failed",
					zap.String("event", e.Value.EventName()), zap.Error(err))
				return nil
			}
			if err := deliver(ctx, obs); err != nil {
				logUtil.GetLogger().Warn("dispatch webhook observation failed",
					zap.String("event", e.Value.EventName()), zap.Error(err))
			}
			return nil
		})
	}
}
