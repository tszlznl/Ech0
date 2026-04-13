package subscriber

import (
	"context"
	"fmt"
	"time"

	contracts "github.com/lin-snow/ech0/internal/event/contracts"
	registry "github.com/lin-snow/ech0/internal/event/registry"
	queueModel "github.com/lin-snow/ech0/internal/model/queue"
)

type DeadLetterProcessor interface {
	HandleDeadLetter(ctx context.Context, deadLetter *queueModel.DeadLetter) error
}

type DeadLetterRepo interface {
	UpdateDeadLetter(ctx context.Context, deadLetter *queueModel.DeadLetter) error
	DeleteDeadLetter(ctx context.Context, id int64) error
}

type DeadLetterResolver struct {
	queueRepo DeadLetterRepo
	processor DeadLetterProcessor
}

func NewDeadLetterResolver(
	queueRepo DeadLetterRepo,
	processor DeadLetterProcessor,
) *DeadLetterResolver {
	return &DeadLetterResolver{
		queueRepo: queueRepo,
		processor: processor,
	}
}

func (dlr *DeadLetterResolver) Handle(ctx context.Context, event contracts.DeadLetterRetriedEvent) error {
	deadLetter := event.DeadLetter

	switch deadLetter.Status {
	case queueModel.DeadLetterStatusPending:
		fallthrough
	case queueModel.DeadLetterStatusFailed:
		if deadLetter.Status == queueModel.DeadLetterStatusFailed && deadLetter.RetryCount >= 3 {
			deadLetter.Status = queueModel.DeadLetterStatusDiscarded
			deadLetter.UpdatedAt = time.Now().UTC().Unix()
			if err := dlr.queueRepo.UpdateDeadLetter(ctx, &deadLetter); err != nil {
				return fmt.Errorf("failed to update dead letter to discarded: %v", err)
			}
			return nil
		}

		deadLetter.Status = queueModel.DeadLetterStatusProcessing
		deadLetter.RetryCount += 1
		deadLetter.UpdatedAt = time.Now().UTC().Unix()
		deadLetter.NextRetry = time.Now().UTC().Add(5 * time.Minute).Unix()

		if err := dlr.queueRepo.UpdateDeadLetter(ctx, &deadLetter); err != nil {
			return fmt.Errorf("failed to update dead letter to processing: %v", err)
		}

		if err := dlr.processDeadLetter(ctx, &deadLetter); err != nil {
			deadLetter.ErrorMsg = err.Error()
			deadLetter.Status = queueModel.DeadLetterStatusFailed
			deadLetter.UpdatedAt = time.Now().UTC().Unix()
			deadLetter.NextRetry = time.Now().UTC().Add(15 * time.Minute).Unix()
			if err := dlr.queueRepo.UpdateDeadLetter(ctx, &deadLetter); err != nil {
				return fmt.Errorf("failed to update dead letter to failed: %v", err)
			}
			return fmt.Errorf("failed to process dead letter: %v", err)
		}

		deadLetter.Status = queueModel.DeadLetterStatusCompleted
		deadLetter.UpdatedAt = time.Now().UTC().Unix()
		if err := dlr.queueRepo.UpdateDeadLetter(ctx, &deadLetter); err != nil {
			return fmt.Errorf("failed to update dead letter to completed: %v", err)
		}
		return nil

	case queueModel.DeadLetterStatusProcessing:
		return nil

	case queueModel.DeadLetterStatusDiscarded:
		if err := dlr.queueRepo.DeleteDeadLetter(ctx, deadLetter.ID); err != nil {
			return fmt.Errorf("failed to delete discarded dead letter: %v", err)
		}
		return nil

	case queueModel.DeadLetterStatusCompleted:
		if err := dlr.queueRepo.DeleteDeadLetter(ctx, deadLetter.ID); err != nil {
			return fmt.Errorf("failed to delete completed dead letter: %v", err)
		}
		return nil

	default:
		deadLetter.Status = queueModel.DeadLetterStatusDiscarded
		if err := dlr.queueRepo.DeleteDeadLetter(ctx, deadLetter.ID); err != nil {
			return fmt.Errorf("failed to delete unknown status dead letter: %v", err)
		}
		return fmt.Errorf("unknown dead letter status: %s", deadLetter.Status)
	}
}

func (dlr *DeadLetterResolver) processDeadLetter(
	ctx context.Context,
	deadLetter *queueModel.DeadLetter,
) error {
	switch deadLetter.Type {
	case queueModel.DeadLetterTypeWebhook:
		return dlr.processor.HandleDeadLetter(ctx, deadLetter)
	default:
		return fmt.Errorf("unknown dead letter type: %s", deadLetter.Type)
	}
}

func (dlr *DeadLetterResolver) Subscriptions() []registry.Subscription {
	return []registry.Subscription{
		registry.TopicSubscription(
			contracts.TopicDeadLetterRetried,
			dlr.Handle,
			registry.DeadLetterSubscribeOptions()...,
		),
	}
}
