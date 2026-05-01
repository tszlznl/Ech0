// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package httpclient

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	contracts "github.com/lin-snow/ech0/internal/event/contracts"
	webhookModel "github.com/lin-snow/ech0/internal/model/webhook"
)

func BuildRequest(wh *webhookModel.Webhook, obs contracts.WebhookObservation) (*http.Request, error) {
	eventID := fmt.Sprintf("%d", time.Now().UTC().UnixNano())
	timestamp := strconv.FormatInt(time.Now().UTC().Unix(), 10)
	headers := make(http.Header)
	headers.Set("Content-Type", "application/json")
	headers.Set("X-Ech0-Event", obs.Topic)
	headers.Set("User-Agent", "Ech0-Webhook-Client")
	headers.Set("X-Ech0-Event-ID", eventID)
	headers.Set("X-Ech0-Timestamp", timestamp)

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

	if wh.Secret != "" {
		signature := buildWebhookSignature(wh.Secret, body)
		headers.Set("X-Ech0-Signature", "sha256="+signature)
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

func SendWithRetry(
	client *http.Client,
	wh *webhookModel.Webhook,
	obs contracts.WebhookObservation,
	maxRetries int,
	initialBackoff time.Duration,
) error {
	return retryWithBackoff(maxRetries, initialBackoff, func() error {
		req, err := BuildRequest(wh, obs)
		if err != nil {
			return err
		}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	})
}

func retryWithBackoff(maxRetries int, initialBackoff time.Duration, fn func() error) error {
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

func buildWebhookSignature(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}
