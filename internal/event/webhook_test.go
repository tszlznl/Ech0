// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package event

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	userModel "github.com/lin-snow/ech0/internal/model/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewWebhookObservation_Success 验证中立观察的构造：topic 透传、payload 正确序列化、
// EventName 取 Go 类型名、metadata 透传、OccurredAt 设为当前 UTC 秒。
func TestNewWebhookObservation_Success(t *testing.T) {
	ev := ResourceUploaded{
		User:     userModel.User{ID: "u-1", Username: "tester"},
		FileName: "pic.png",
		URL:      "https://cdn.example/pic.png",
		Size:     2048,
		Type:     "image/png",
		Key:      "uploads/pic.png",
	}
	meta := map[string]string{"trace": "abc123", "actor": "u-1"}

	before := time.Now().UTC().Unix()
	obs, err := NewWebhookObservation(ev.EventName(), ev, meta)
	after := time.Now().UTC().Unix()

	require.NoError(t, err)

	t.Run("topic is the stable event name", func(t *testing.T) {
		assert.Equal(t, "resource.uploaded", obs.Topic)
	})

	t.Run("event_name is the Go type name", func(t *testing.T) {
		// 注意：EventName 字段是 Go 类型名（ResourceUploaded），与 Topic（稳定名）不同。
		assert.Equal(t, "ResourceUploaded", obs.EventName)
	})

	t.Run("metadata passes through unchanged", func(t *testing.T) {
		assert.Equal(t, meta, obs.Metadata)
	})

	t.Run("occurred_at within call window", func(t *testing.T) {
		assert.GreaterOrEqual(t, obs.OccurredAt, before)
		assert.LessOrEqual(t, obs.OccurredAt, after)
	})

	t.Run("payload round-trips to original fields", func(t *testing.T) {
		var got map[string]any
		require.NoError(t, json.Unmarshal(obs.Payload, &got))
		assert.Equal(t, "pic.png", got["FileName"])
		assert.Equal(t, "https://cdn.example/pic.png", got["URL"])
		assert.EqualValues(t, 2048, got["Size"])
		assert.Equal(t, "image/png", got["Type"])
		assert.Contains(t, got, "User")
	})
}

// TestNewWebhookObservation_ResourceUploadedKeyOmitted 回归：ResourceUploaded.Key 带 json:"-"，
// 仅用于发布时的局部有序，绝不能出现在 webhook 载荷中（否则泄露存储路径并破坏载荷契约）。
func TestNewWebhookObservation_ResourceUploadedKeyOmitted(t *testing.T) {
	const secretKey = "private/storage/secret-key.bin"
	ev := ResourceUploaded{
		User:     userModel.User{ID: "u-1"},
		FileName: "a.png",
		URL:      "https://cdn.example/a.png",
		Size:     1,
		Type:     "image/png",
		Key:      secretKey,
	}

	obs, err := NewWebhookObservation(ev.EventName(), ev, nil)
	require.NoError(t, err)

	// 原始字节中不应出现存储 key（既不作为字段名，也不作为值）。
	assert.NotContains(t, string(obs.Payload), secretKey)

	var got map[string]any
	require.NoError(t, json.Unmarshal(obs.Payload, &got))
	assert.NotContains(t, got, "Key", "Key (json:\"-\") must not leak into payload")
}

// TestNewWebhookObservation_NilMetadata 验证 nil metadata 透传为 nil（omitempty 时省略）。
func TestNewWebhookObservation_NilMetadata(t *testing.T) {
	obs, err := NewWebhookObservation("echo.created", EchoCreated{}, nil)
	require.NoError(t, err)
	assert.Nil(t, obs.Metadata)
}

// TestNewWebhookObservation_TopicIsCallerSupplied 验证 Topic 完全取调用方入参，不被内部覆写。
func TestNewWebhookObservation_TopicIsCallerSupplied(t *testing.T) {
	obs, err := NewWebhookObservation("custom.topic", EchoCreated{}, nil)
	require.NoError(t, err)
	assert.Equal(t, "custom.topic", obs.Topic)
	// EventName 字段仍是 Go 类型名，与 Topic 解耦。
	assert.Equal(t, "EchoCreated", obs.EventName)
}

// TestNewWebhookObservation_MarshalError 验证不可序列化的 payload 返回错误且观察为零值。
func TestNewWebhookObservation_MarshalError(t *testing.T) {
	// chan 无法被 json 序列化，json.Marshal 返回 *json.UnsupportedTypeError。
	obs, err := NewWebhookObservation("bad", make(chan int), nil)
	require.Error(t, err)
	assert.Equal(t, WebhookObservation{}, obs, "error path must return zero observation")
}

// TestEventNameOf 锁定类型名推导：值类型取 Name，指针解引用，nil 返回空，匿名类型回退到 String。
func TestEventNameOf(t *testing.T) {
	ev := EchoCreated{}

	t.Run("value type yields struct name", func(t *testing.T) {
		assert.Equal(t, "EchoCreated", eventNameOf(ev))
	})

	t.Run("pointer is dereferenced to elem name", func(t *testing.T) {
		assert.Equal(t, "EchoCreated", eventNameOf(&ev))
	})

	t.Run("nil yields empty string", func(t *testing.T) {
		assert.Equal(t, "", eventNameOf(nil))
	})

	t.Run("anonymous struct falls back to String", func(t *testing.T) {
		anon := struct{ X int }{X: 1}
		got := eventNameOf(anon)
		assert.NotEmpty(t, got)
		assert.Equal(t, reflect.TypeOf(anon).String(), got)
	})
}
