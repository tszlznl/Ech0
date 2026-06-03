// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package webhook

import (
	"net/http"
	"time"

	"github.com/lin-snow/ech0/internal/event"
	webhookModel "github.com/lin-snow/ech0/internal/model/webhook"
	"github.com/lin-snow/ech0/internal/util/egress"
)

const (
	defaultWebhookTimeout = 5 * time.Second

	deliverMaxRetries = 3
	deliverBackoff    = 500 * time.Millisecond
	testMaxRetries    = 2
	testBackoff       = 300 * time.Millisecond
)

// Sender 是 webhook 的唯一出网出口：持有出网 HTTP client，负责签名构造 + 重试发送。
// 正式投递（Dispatcher）与连通性测试（设置页 TestWebhook）共用它，避免 client 构造、
// 超时、重试参数在两处各写一份而漂移。
type Sender struct {
	client *http.Client
}

func NewSender() *Sender {
	return &Sender{
		client: egress.NewClient(egress.Guard(), egress.Timeout(defaultWebhookTimeout)),
	}
}

// Deliver 投递一次正式事件观察。即时重试仍失败则由调用方（Dispatcher）记录失败状态，不再补投。
func (s *Sender) Deliver(wh *webhookModel.Webhook, obs event.WebhookObservation) error {
	return sendWithRetry(s.client, wh, obs, deliverMaxRetries, deliverBackoff)
}

// SendTest 构造一次连通性测试观察并发送，供设置页 TestWebhook 复用。
func (s *Sender) SendTest(wh *webhookModel.Webhook) error {
	obs, err := event.NewWebhookObservation("webhook.test", map[string]any{
		"message": "webhook connectivity test from ech0",
		"webhook": wh.Name,
		"time":    time.Now().UTC().Format(time.RFC3339),
	}, map[string]string{"source": "setting.test"})
	if err != nil {
		return err
	}
	return sendWithRetry(s.client, wh, obs, testMaxRetries, testBackoff)
}
