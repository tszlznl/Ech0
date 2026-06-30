// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package webhook

import (
	"encoding/json"
	"io"
	"strconv"
	"testing"

	"github.com/lin-snow/ech0/internal/event"
	webhookModel "github.com/lin-snow/ech0/internal/model/webhook"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuildWebhookSignature 校验 HMAC-SHA256 签名：对已知输入稳定、密钥/载荷不同则签名不同。
func TestBuildWebhookSignature(t *testing.T) {
	t.Run("known vector is stable", func(t *testing.T) {
		// RFC/Wikipedia 规范 HMAC-SHA256 测试向量：
		// HMAC-SHA256("key", "The quick brown fox jumps over the lazy dog")
		const (
			secret  = "key"
			payload = "The quick brown fox jumps over the lazy dog"
			want    = "f7bc83f430538424b13298e6aa6fb143ef4d59a14946175997479dbc2d1a3cd8"
		)
		got := buildWebhookSignature(secret, []byte(payload))
		assert.Equal(t, want, got, "签名必须匹配已知 HMAC-SHA256 向量")
		// 长度恒为 32 字节 hex = 64 字符。
		assert.Len(t, got, 64)
	})

	t.Run("deterministic across calls", func(t *testing.T) {
		secret := "topsecret"
		payload := []byte(`{"topic":"echo.created","n":1}`)
		first := buildWebhookSignature(secret, payload)
		second := buildWebhookSignature(secret, payload)
		assert.Equal(t, first, second, "相同密钥+相同载荷必须产生相同签名")
	})

	t.Run("different secret yields different signature", func(t *testing.T) {
		payload := []byte(`{"topic":"echo.created"}`)
		a := buildWebhookSignature("secret-a", payload)
		b := buildWebhookSignature("secret-b", payload)
		assert.NotEqual(t, a, b, "不同密钥应产生不同签名")
		assert.Len(t, a, 64)
		assert.Len(t, b, 64)
	})

	t.Run("different payload yields different signature", func(t *testing.T) {
		secret := "same-secret"
		a := buildWebhookSignature(secret, []byte(`{"v":1}`))
		b := buildWebhookSignature(secret, []byte(`{"v":2}`))
		assert.NotEqual(t, a, b, "不同载荷应产生不同签名")
	})

	t.Run("empty payload still produces 64-char hex", func(t *testing.T) {
		got := buildWebhookSignature("k", nil)
		assert.Len(t, got, 64)
		// 空载荷下也应稳定。
		assert.Equal(t, got, buildWebhookSignature("k", []byte{}))
	})
}

// newObs 构造一个测试用的中立观察。
func newObs(t *testing.T) event.WebhookObservation {
	t.Helper()
	return event.WebhookObservation{
		Topic:      "echo.created",
		EventName:  "EchoCreated",
		Payload:    json.RawMessage(`{"id":42,"content":"hello"}`),
		Metadata:   map[string]string{"source": "test"},
		OccurredAt: 1700000000,
	}
}

// readBody 通过 GetBody 取一份新鲜的 body 拷贝，避免消费 req.Body。
func readBody(t *testing.T, getBody func() (io.ReadCloser, error)) []byte {
	t.Helper()
	require.NotNil(t, getBody, "buildRequest 必须设置 GetBody")
	rc, err := getBody()
	require.NoError(t, err)
	defer func() { _ = rc.Close() }()
	b, err := io.ReadAll(rc)
	require.NoError(t, err)
	return b
}

// TestBuildRequest_HeaderContract 校验请求构造的 header 契约与方法/URL。
func TestBuildRequest_HeaderContract(t *testing.T) {
	wh := &webhookModel.Webhook{
		ID:     "wh-1",
		Name:   "demo",
		URL:    "https://example.com/hook",
		Secret: "s3cr3t",
	}
	obs := newObs(t)

	req, err := buildRequest(wh, obs)
	require.NoError(t, err)
	require.NotNil(t, req)

	assert.Equal(t, "POST", req.Method)
	assert.Equal(t, wh.URL, req.URL.String())

	// 固定 header 契约。
	assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
	assert.Equal(t, obs.Topic, req.Header.Get("X-Ech0-Event"))
	assert.Equal(t, "Ech0-Webhook-Client", req.Header.Get("User-Agent"))

	// 动态但必须存在且形态合理的 header。
	eventID := req.Header.Get("X-Ech0-Event-ID")
	require.NotEmpty(t, eventID, "X-Ech0-Event-ID 必须存在")
	_, err = strconv.ParseInt(eventID, 10, 64)
	assert.NoError(t, err, "X-Ech0-Event-ID 应为纳秒时间戳数字")

	ts := req.Header.Get("X-Ech0-Timestamp")
	require.NotEmpty(t, ts, "X-Ech0-Timestamp 必须存在")
	_, err = strconv.ParseInt(ts, 10, 64)
	assert.NoError(t, err, "X-Ech0-Timestamp 应为 unix 秒数字")
}

// TestBuildRequest_SignatureHeader 校验：有密钥时签名头存在且与 body 一致，无密钥时不带签名头。
func TestBuildRequest_SignatureHeader(t *testing.T) {
	obs := newObs(t)

	t.Run("with secret sets sha256 signature over body", func(t *testing.T) {
		wh := &webhookModel.Webhook{URL: "https://example.com/hook", Secret: "abc123"}
		req, err := buildRequest(wh, obs)
		require.NoError(t, err)

		sig := req.Header.Get("X-Ech0-Signature")
		require.NotEmpty(t, sig)
		assert.True(t, len(sig) > len("sha256="), "签名头应带 sha256= 前缀")
		assert.Equal(t, "sha256=", sig[:7], "签名头前缀必须是 sha256=")

		// 签名必须覆盖实际发送的 body。
		body := readBody(t, req.GetBody)
		want := "sha256=" + buildWebhookSignature(wh.Secret, body)
		assert.Equal(t, want, sig, "签名头必须等于对 body 的 HMAC")
	})

	t.Run("without secret omits signature header", func(t *testing.T) {
		wh := &webhookModel.Webhook{URL: "https://example.com/hook", Secret: ""}
		req, err := buildRequest(wh, obs)
		require.NoError(t, err)
		assert.Empty(t, req.Header.Get("X-Ech0-Signature"), "无密钥不应带签名头")
	})
}

// TestBuildRequest_BodyShape 校验 body 的 JSON 结构与字段映射。
func TestBuildRequest_BodyShape(t *testing.T) {
	wh := &webhookModel.Webhook{URL: "https://example.com/hook"}
	obs := newObs(t)

	req, err := buildRequest(wh, obs)
	require.NoError(t, err)

	body := readBody(t, req.GetBody)

	var decoded struct {
		Topic      string            `json:"topic"`
		EventName  string            `json:"event_name"`
		PayloadRaw json.RawMessage   `json:"payload_raw"`
		Metadata   map[string]string `json:"metadata"`
		OccurredAt int64             `json:"occurred_at"`
	}
	require.NoError(t, json.Unmarshal(body, &decoded))

	assert.Equal(t, obs.Topic, decoded.Topic)
	assert.Equal(t, obs.EventName, decoded.EventName)
	assert.JSONEq(t, string(obs.Payload), string(decoded.PayloadRaw))
	assert.Equal(t, obs.Metadata, decoded.Metadata)
	assert.Equal(t, obs.OccurredAt, decoded.OccurredAt)

	// GetBody 可重复读取且内容一致（重试时复用）。
	again := readBody(t, req.GetBody)
	assert.Equal(t, body, again, "GetBody 应可重复返回相同 body")
}

// TestBuildRequest_InvalidURL 校验非法 URL 时返回错误。
func TestBuildRequest_InvalidURL(t *testing.T) {
	wh := &webhookModel.Webhook{URL: "://bad-url"}
	_, err := buildRequest(wh, newObs(t))
	assert.Error(t, err, "非法 URL 应导致 http.NewRequest 报错")
}
