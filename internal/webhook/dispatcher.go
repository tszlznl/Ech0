// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/lin-snow/ech0/internal/async"
	"github.com/lin-snow/ech0/internal/config"
	contracts "github.com/lin-snow/ech0/internal/event/contracts"
	queueModel "github.com/lin-snow/ech0/internal/model/queue"
	webhookModel "github.com/lin-snow/ech0/internal/model/webhook"
	"github.com/lin-snow/ech0/internal/transaction"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	webhookclient "github.com/lin-snow/ech0/internal/webhook/infra/httpclient"
	"go.uber.org/zap"
)

const defaultWebhookTimeout = 5 * time.Second

type WebhookStore interface {
	ListActiveWebhooks(ctx context.Context) ([]webhookModel.Webhook, error)
	UpdateWebhookDeliveryStatus(
		ctx context.Context,
		id string,
		status string,
		lastTrigger int64,
	) error
}

type DeadLetterStore interface {
	SaveDeadLetter(ctx context.Context, deadLetter *queueModel.DeadLetter) error
}

type Dispatcher struct {
	client     *http.Client
	repo       WebhookStore
	pool       *async.WorkerPool
	queueRepo  DeadLetterStore
	transactor transaction.Transactor
}

func NewDispatcher(
	repo WebhookStore,
	queueRepo DeadLetterStore,
	tx transaction.Transactor,
) *Dispatcher {
	return &Dispatcher{
		repo:      repo,
		queueRepo: queueRepo,
		client:    webhookclient.NewSafeHTTPClient(defaultWebhookTimeout),
		pool: async.NewWorkerPool(
			config.Config().Event.WebhookPoolWorkers,
			config.Config().Event.WebhookPoolQueue,
		),
		transactor: tx,
	}
}

func (wd *Dispatcher) HandleObservation(ctx context.Context, obs contracts.WebhookObservation) error {
	webhooks, err := wd.repo.ListActiveWebhooks(ctx)
	if err != nil {
		return err
	}
	for _, wh := range webhooks {
		wh := wh
		wd.pool.Submit(func() error {
			wd.Dispatch(ctx, &wh, obs)
			return nil
		})
	}

	return nil
}

func (wd *Dispatcher) Dispatch(ctx context.Context, wh *webhookModel.Webhook, obs contracts.WebhookObservation) {
	triggerAt := time.Now().UTC().Unix()
	err := webhookclient.SendWithRetry(wd.client, wh, obs, 3, 500*time.Millisecond)
	if err != nil {
		wd.updateWebhookStatus(ctx, wh.ID, "failed", triggerAt)
		logUtil.GetLogger().Error("Webhook Handle Failed", zap.String("name", wh.Name), zap.String("url", wh.URL), zap.Error(err))

		payloadData := contracts.WebhookReplayPayload{
			Webhook: *wh,
			Event:   obs,
		}
		payload, _ := json.Marshal(payloadData)

		var deadLetter queueModel.DeadLetter
		deadLetter.SetType(queueModel.DeadLetterTypeWebhook)
		deadLetter.Payload = payload
		deadLetter.ErrorMsg = err.Error()
		deadLetter.RetryCount = 0
		deadLetter.NextRetry = time.Now().UTC().Add(6 * time.Hour).Unix()
		deadLetter.CreatedAt = time.Now().UTC().Unix()
		deadLetter.UpdatedAt = time.Now().UTC().Unix()
		deadLetter.Status = queueModel.DeadLetterStatusPending

		if err := wd.transactor.Run(ctx, func(ctx context.Context) error {
			return wd.queueRepo.SaveDeadLetter(ctx, &deadLetter)
		}); err != nil {
			logUtil.GetLogger().Error("Failed to save dead letter", zap.Error(err))
		}
		return
	}
	wd.updateWebhookStatus(ctx, wh.ID, "success", triggerAt)
}

func (wd *Dispatcher) Wait() {
	wd.pool.Wait()
}

func (wd *Dispatcher) Stop() {
	wd.pool.Stop()
}

func (wd *Dispatcher) HandleDeadLetter(
	ctx context.Context,
	deadLetter *queueModel.DeadLetter,
) error {
	var payload contracts.WebhookReplayPayload
	if err := json.Unmarshal(deadLetter.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal dead letter payload: %w", err)
	}
	webhook := payload.Webhook
	obs := payload.Event
	triggerAt := time.Now().UTC().Unix()

	err := webhookclient.SendWithRetry(wd.client, &webhook, obs, 3, 500*time.Millisecond)
	if err != nil {
		wd.updateWebhookStatus(ctx, webhook.ID, "failed", triggerAt)
		return err
	}
	wd.updateWebhookStatus(ctx, webhook.ID, "success", triggerAt)
	return nil
}

func (wd *Dispatcher) updateWebhookStatus(
	ctx context.Context,
	webhookID string,
	status string,
	triggerAt int64,
) {
	if webhookID == "" {
		return
	}
	if err := wd.repo.UpdateWebhookDeliveryStatus(ctx, webhookID, status, triggerAt); err != nil {
		logUtil.GetLogger().Warn(
			"update webhook delivery status failed",
			zap.String("webhook_id", webhookID),
			zap.String("status", status),
			zap.Error(err),
		)
	}
}
