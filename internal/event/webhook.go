// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package event

import (
	"encoding/json"
	"reflect"
	"time"
)

// WebhookObservation 是一次“已发生事件”的中立快照，供 webhook 分发与重放使用。
type WebhookObservation struct {
	Topic      string            `json:"topic"`
	EventName  string            `json:"event_name"`
	Payload    json.RawMessage   `json:"payload"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	OccurredAt int64             `json:"occurred_at"`
}

// NewWebhookObservation 把一个事件序列化为中立观察。topic 取事件的稳定名（EventName）。
func NewWebhookObservation(topic string, payload any, metadata map[string]string) (WebhookObservation, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return WebhookObservation{}, err
	}
	return WebhookObservation{
		Topic:      topic,
		EventName:  eventNameOf(payload),
		Payload:    raw,
		Metadata:   metadata,
		OccurredAt: time.Now().UTC().Unix(),
	}, nil
}

func eventNameOf(payload any) string {
	if payload == nil {
		return ""
	}
	t := reflect.TypeOf(payload)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Name() != "" {
		return t.Name()
	}
	return t.String()
}
