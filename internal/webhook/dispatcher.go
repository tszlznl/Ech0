// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package webhook

import (
	"context"
	"net/http"
	"time"

	"github.com/lin-snow/ech0/internal/config"
	"github.com/lin-snow/ech0/internal/event"
	webhookModel "github.com/lin-snow/ech0/internal/model/webhook"
	asyncUtil "github.com/lin-snow/ech0/internal/util/async"
	"github.com/lin-snow/ech0/internal/util/egress"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
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

type Dispatcher struct {
	client *http.Client
	repo   WebhookStore
	pool   *asyncUtil.WorkerPool
}

func NewDispatcher(repo WebhookStore) *Dispatcher {
	return &Dispatcher{
		repo:   repo,
		client: egress.NewClient(egress.Guard(), egress.Timeout(defaultWebhookTimeout)),
		pool: asyncUtil.NewWorkerPool(
			config.Config().Event.WebhookPoolWorkers,
			config.Config().Event.WebhookPoolQueue,
		),
	}
}

func (wd *Dispatcher) HandleObservation(ctx context.Context, obs event.WebhookObservation) error {
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

func (wd *Dispatcher) Dispatch(ctx context.Context, wh *webhookModel.Webhook, obs event.WebhookObservation) {
	triggerAt := time.Now().UTC().Unix()
	if err := SendWithRetry(wd.client, wh, obs, 3, 500*time.Millisecond); err != nil {
		wd.updateWebhookStatus(ctx, wh.ID, "failed", triggerAt)
		logUtil.GetLogger().Error("Webhook Handle Failed", zap.String("name", wh.Name), zap.String("url", wh.URL), zap.Error(err))
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
