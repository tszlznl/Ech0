package subscriber

import (
	"context"
	"fmt"
	"time"

	contracts "github.com/lin-snow/ech0/internal/event/contracts"
	queueModel "github.com/lin-snow/ech0/internal/model/queue"
	queueRepository "github.com/lin-snow/ech0/internal/repository/queue"
)

type DeadLetterProcessor interface {
	HandleDeadLetter(ctx context.Context, deadLetter *queueModel.DeadLetter) error
}

type DeadLetterResolver struct {
	queueRepo queueRepository.QueueRepositoryInterface
	processor DeadLetterProcessor
}

func NewDeadLetterResolver(
	queueRepo queueRepository.QueueRepositoryInterface,
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
		deadLetter.Status = queueModel.DeadLetterStatusProcessing
		deadLetter.RetryCount += 1
		deadLetter.UpdatedAt = time.Now().UTC()
		deadLetter.NextRetry = time.Now().UTC().Add(6 * time.Hour)

		if err := dlr.queueRepo.UpdateDeadLetter(ctx, &deadLetter); err != nil {
			return fmt.Errorf("failed to update dead letter to processing: %v", err)
		}

		if err := dlr.processDeadLetter(ctx, &deadLetter); err != nil {
			deadLetter.ErrorMsg = err.Error()
			deadLetter.Status = queueModel.DeadLetterStatusFailed
			deadLetter.UpdatedAt = time.Now().UTC()
			if err := dlr.queueRepo.UpdateDeadLetter(ctx, &deadLetter); err != nil {
				return fmt.Errorf("failed to update dead letter to failed: %v", err)
			}
			return fmt.Errorf("failed to process dead letter: %v", err)
		}

		deadLetter.Status = queueModel.DeadLetterStatusCompleted
		deadLetter.UpdatedAt = time.Now().UTC()
		if err := dlr.queueRepo.UpdateDeadLetter(ctx, &deadLetter); err != nil {
			return fmt.Errorf("failed to update dead letter to completed: %v", err)
		}
		return nil

	case queueModel.DeadLetterStatusProcessing:
		return nil

	case queueModel.DeadLetterStatusFailed:
		if deadLetter.RetryCount <= 3 {
			deadLetter.Status = queueModel.DeadLetterStatusPending
		} else {
			deadLetter.Status = queueModel.DeadLetterStatusDiscarded
		}
		deadLetter.UpdatedAt = time.Now().UTC()
		if err := dlr.queueRepo.UpdateDeadLetter(ctx, &deadLetter); err != nil {
			return fmt.Errorf("failed to update dead letter: %v", err)
		}
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
