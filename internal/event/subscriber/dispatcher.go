package subscriber

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/lin-snow/ech0/internal/async"
	"github.com/lin-snow/ech0/internal/config"
	contracts "github.com/lin-snow/ech0/internal/event/contracts"
	queueModel "github.com/lin-snow/ech0/internal/model/queue"
	webhookModel "github.com/lin-snow/ech0/internal/model/webhook"
	"github.com/lin-snow/ech0/internal/transaction"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"go.uber.org/zap"
)

type WebhookStore interface {
	ListActiveWebhooks(ctx context.Context) ([]webhookModel.Webhook, error)
}

type DeadLetterStore interface {
	SaveDeadLetter(ctx context.Context, deadLetter *queueModel.DeadLetter) error
}

type WebhookDispatcher struct {
	client     *http.Client
	repo       WebhookStore
	pool       *async.WorkerPool
	queueRepo  DeadLetterStore
	transactor transaction.Transactor
}

func NewWebhookDispatcher(
	repo WebhookStore,
	queueRepo DeadLetterStore,
	tx transaction.Transactor,
) *WebhookDispatcher {
	return &WebhookDispatcher{
		repo:      repo,
		queueRepo: queueRepo,
		client: &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     30 * time.Second,
			},
		},
		pool: async.NewWorkerPool(
			config.Config().Event.WebhookPoolWorkers,
			config.Config().Event.WebhookPoolQueue,
		),
		transactor: tx,
	}
}

func (wd *WebhookDispatcher) HandleObservation(ctx context.Context, obs contracts.WebhookObservation) error {
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

func (wd *WebhookDispatcher) Dispatch(ctx context.Context, wh *webhookModel.Webhook, obs contracts.WebhookObservation) {
	err := wd.retryWithBackoff(3, 500*time.Millisecond, func() error {
		req, err := wd.buildRequest(wh, obs)
		if err != nil {
			return err
		}

		resp, err := wd.client.Do(req)
		if err != nil {
			return err
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	})
	if err != nil {
		logUtil.GetLogger().Error("Webhook Handle Failed", zap.String("name", wh.Name), zap.String("url", wh.URL))

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
		deadLetter.NextRetry = time.Now().UTC().Add(6 * time.Hour)
		deadLetter.CreatedAt = time.Now().UTC()
		deadLetter.UpdatedAt = time.Now().UTC()
		deadLetter.Status = queueModel.DeadLetterStatusPending

		if err := wd.transactor.Run(ctx, func(ctx context.Context) error {
			return wd.queueRepo.SaveDeadLetter(ctx, &deadLetter)
		}); err != nil {
			logUtil.GetLogger().Error("Failed to save dead letter", zap.String("error", err.Error()))
		}
	}
}

func (wd *WebhookDispatcher) buildRequest(
	wh *webhookModel.Webhook,
	obs contracts.WebhookObservation,
) (*http.Request, error) {
	eventID := fmt.Sprintf("%d", time.Now().UTC().UnixNano())
	headers := make(http.Header)
	headers.Set("Content-Type", "application/json")
	headers.Set("X-Ech0-Event", obs.Topic)
	headers.Set("User-Agent", "Ech0-Webhook-Client")
	headers.Set("E-Ech0-Event-ID", eventID)

	body, err := json.Marshal(map[string]any{
		"topic":       obs.Topic,
		"event_name":  obs.EventName,
		"payload_raw": obs.Payload,
		"metadata":    obs.Metadata,
		"occurred_at": obs.OccurredAt,
	})
	if err != nil {
		return nil, err
	}
	bodyReader := io.NopCloser(bytes.NewReader(body))

	req, err := http.NewRequest("POST", wh.URL, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header = headers
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(body)), nil
	}

	return req, nil
}

func (wd *WebhookDispatcher) retryWithBackoff(
	maxRetries int,
	initialBackoff time.Duration,
	fn func() error,
) error {
	var err error
	delay := initialBackoff
	for i := 0; i < maxRetries; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		time.Sleep(delay)
		delay *= 2
	}
	return err
}

func (wd *WebhookDispatcher) Wait() {
	wd.pool.Wait()
}

func (wd *WebhookDispatcher) Stop() {
	wd.pool.Stop()
}

func (wd *WebhookDispatcher) HandleDeadLetter(
	ctx context.Context,
	deadLetter *queueModel.DeadLetter,
) error {
	var payload contracts.WebhookReplayPayload
	if err := json.Unmarshal(deadLetter.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal dead letter payload: %w", err)
	}
	webhook := payload.Webhook
	obs := payload.Event

	err := wd.retryWithBackoff(3, 500*time.Millisecond, func() error {
		req, err := wd.buildRequest(&webhook, obs)
		if err != nil {
			return err
		}

		resp, err := wd.client.Do(req)
		if err != nil {
			return err
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	})
	if err != nil {
		return err
	}
	return nil
}
