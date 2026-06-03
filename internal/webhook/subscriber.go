// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package webhook

import (
	"context"

	"github.com/lin-snow/ech0/internal/event"
	eventbus "github.com/lin-snow/ech0/internal/event/bus"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"go.uber.org/zap"
)

// Registrations 让 Dispatcher 作为普通事件订阅者自注册：为每个可观测事件登记一条同步订阅，
// 把强类型事件转为中立 WebhookObservation 后投递。并发 / 重试由 Dispatcher 的 worker pool 承担，
// 故此处同步即可。新增可观测事件时，记得在此追加对应的 observe 行。
func (wd *Dispatcher) Registrations() []eventbus.Registration {
	return []eventbus.Registration{
		observe[event.UserCreated](wd.HandleObservation),
		observe[event.UserUpdated](wd.HandleObservation),
		observe[event.UserDeleted](wd.HandleObservation),
		observe[event.EchoCreated](wd.HandleObservation),
		observe[event.EchoUpdated](wd.HandleObservation),
		observe[event.EchoDeleted](wd.HandleObservation),
		observe[event.CommentCreated](wd.HandleObservation),
		observe[event.CommentStatusUpdated](wd.HandleObservation),
		observe[event.CommentDeleted](wd.HandleObservation),
		observe[event.ResourceUploaded](wd.HandleObservation),
		observe[event.SystemSnapshot](wd.HandleObservation),
		observe[event.SystemExport](wd.HandleObservation),
		observe[event.UpdateSnapshotSchedule](wd.HandleObservation),
	}
}

// observe 构造单个事件类型的 webhook 观察订阅（同步）。它需要 busen 信封的 Meta（source 等元数据），
// 故走 eventbus.OnWithMeta 而非 On —— On 只透传 Value。
func observe[T event.Named](
	deliver func(context.Context, event.WebhookObservation) error,
) eventbus.Registration {
	return eventbus.OnWithMeta(func(ctx context.Context, v T, meta map[string]string) error {
		obs, err := event.NewWebhookObservation(v.EventName(), v, meta)
		if err != nil {
			logUtil.GetLogger().Warn("build webhook observation failed",
				zap.String("event", v.EventName()), zap.Error(err))
			return nil
		}
		if err := deliver(ctx, obs); err != nil {
			logUtil.GetLogger().Warn("dispatch webhook observation failed",
				zap.String("event", v.EventName()), zap.Error(err))
		}
		return nil
	})
}
