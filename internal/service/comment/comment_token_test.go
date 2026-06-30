// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/lin-snow/ech0/internal/test/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 这些是纯函数（未导出）的 in-package 测试：只依赖全局 JWT 密钥，不需要 mock。

func TestFormTokenRoundTrip(t *testing.T) {
	helpers.SetJWTSecret(t, "round-trip-secret")
	s := &CommentService{}

	const ip = "203.0.113.9"
	// 5s 之前签发：大于 minSubmitMS、远小于 maxFormTokenHours，处于有效窗口内。
	issuedAt := time.Now().UnixMilli() - 5000

	token := s.signFormToken(ip, issuedAt)
	require.NotEmpty(t, token)
	require.NoError(t, s.verifyFormToken(ip, token), "freshly signed token must verify")
}

func TestVerifyFormToken_Rejects(t *testing.T) {
	helpers.SetJWTSecret(t, "reject-secret")
	s := &CommentService{}

	const ip = "198.51.100.7"
	now := time.Now().UnixMilli()

	// manualSign 用指定密钥手工复刻签名算法，用于构造「换密钥伪造」的 token。
	manualSign := func(clientIP string, issuedAt int64, secret string) string {
		mac := hmac.New(sha256.New, []byte(secret))
		_, _ = fmt.Fprintf(mac, "%s:%d", clientIP, issuedAt)
		sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
		return fmt.Sprintf("%d.%s", issuedAt, sig)
	}

	// tamper 翻转 token 最后一个字符，破坏签名段。
	tamper := func(token string) string {
		b := []byte(token)
		last := len(b) - 1
		if b[last] == 'A' {
			b[last] = 'B'
		} else {
			b[last] = 'A'
		}
		return string(b)
	}

	cases := []struct {
		name            string
		token           string
		verifyIP        string
		wantErrContains string
	}{
		{"empty token", "", ip, "token invalid"},
		{"single segment", "noseparator", ip, "token invalid"},
		{"three segments", "a.b.c", ip, "token invalid"},
		{"non-numeric issuedAt", "abc.sig", ip, "invalid syntax"},
		{"submit too fast", s.signFormToken(ip, now-500), ip, "submit too fast"},
		{
			"expired",
			s.signFormToken(ip, now-(maxFormTokenHours*3600*1000+5000)),
			ip,
			"token expired",
		},
		{"tampered signature", tamper(s.signFormToken(ip, now-5000)), ip, "token sign mismatch"},
		{"client ip mismatch", s.signFormToken("10.0.0.1", now-5000), ip, "token sign mismatch"},
		{"forged with wrong key", manualSign(ip, now-5000, "a-totally-different-secret"), ip, "token sign mismatch"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := s.verifyFormToken(tc.verifyIP, tc.token)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErrContains)
		})
	}
}

func TestComputeHMAC(t *testing.T) {
	helpers.SetJWTSecret(t, "hmac-secret")
	s := &CommentService{}

	a := s.computeHMAC("1.2.3.4:1700000000000")
	b := s.computeHMAC("1.2.3.4:1700000000000")
	assert.Equal(t, a, b, "same payload + key must be deterministic")
	assert.NotEmpty(t, a)

	c := s.computeHMAC("1.2.3.4:1700000000001")
	assert.NotEqual(t, a, c, "different payload must yield different mac")

	// 输出必须是合法的 base64 RawURL 编码。
	_, err := base64.RawURLEncoding.DecodeString(a)
	assert.NoError(t, err)
}
