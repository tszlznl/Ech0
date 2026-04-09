package registry

import (
	"context"

	busen "github.com/lin-snow/Busen"
	contracts "github.com/lin-snow/ech0/internal/event/contracts"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
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
		webhookTopicSubscription[contracts.SystemBackupEvent](observer, contracts.TopicSystemBackup),
		webhookTopicSubscription[contracts.SystemExportEvent](observer, contracts.TopicSystemExport),
		webhookTopicSubscription[contracts.UpdateBackupScheduleEvent](observer, contracts.TopicBackupScheduleUpdate),
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
