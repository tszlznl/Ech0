// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package registry

import (
	"context"

	contracts "github.com/lin-snow/ech0/internal/event/contracts"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"github.com/lin-snow/ech0/pkg/busen"
	"go.uber.org/zap"
)

func webhookSubscriptions(observer WebhookObserver) []Subscription {
	return []Subscription{
		webhookTopicSubscription[contracts.UserCreatedEvent](observer, contracts.TopicUserCreated),
		webhookTopicSubscription[contracts.UserUpdatedEvent](observer, contracts.TopicUserUpdated),
		webhookTopicSubscription[contracts.UserDeletedEvent](observer, contracts.TopicUserDeleted),
		webhookTopicSubscription[contracts.EchoCreatedEvent](observer, contracts.TopicEchoCreated),
		webhookTopicSubscription[contracts.EchoUpdatedEvent](observer, contracts.TopicEchoUpdated),
		webhookTopicSubscription[contracts.EchoDeletedEvent](observer, contracts.TopicEchoDeleted),
		webhookTopicSubscription[contracts.CommentCreatedEvent](observer, contracts.TopicCommentCreated),
		webhookTopicSubscription[contracts.CommentStatusUpdatedEvent](observer, contracts.TopicCommentStatusUpdated),
		webhookTopicSubscription[contracts.CommentDeletedEvent](observer, contracts.TopicCommentDeleted),
		webhookTopicSubscription[contracts.ResourceUploadedEvent](observer, contracts.TopicResourceUploaded),
		webhookTopicSubscription[contracts.SystemSnapshotEvent](observer, contracts.TopicSystemSnapshot),
		webhookTopicSubscription[contracts.SystemExportEvent](observer, contracts.TopicSystemExport),
		webhookTopicSubscription[contracts.UpdateSnapshotScheduleEvent](observer, contracts.TopicSnapshotScheduleUpdate),
	}
}

func webhookTopicSubscription[T any](observer WebhookObserver, topic string) Subscription {
	return Subscription{
		register: func(bus *busen.Bus) (func(), error) {
			return busen.SubscribeTopic(bus, topic, func(ctx context.Context, e busen.Event[T]) error {
				obs, err := contracts.NewWebhookObservation(e.Topic, e.Value, e.Meta)
				if err != nil {
					logUtil.GetLogger().Warn("build webhook observation failed",
						zap.String("topic", e.Topic),
						zap.Error(err))
					return nil
				}
				if err := observer.HandleObservation(ctx, obs); err != nil {
					logUtil.GetLogger().Warn("dispatch webhook observation failed",
						zap.String("topic", e.Topic),
						zap.Error(err))
				}
				return nil
			})
		},
	}
}
