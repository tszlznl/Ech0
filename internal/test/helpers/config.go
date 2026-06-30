// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package helpers

import (
	"testing"

	"github.com/lin-snow/ech0/internal/config"
)

// SetJWTSecret 覆写测试期的全局 JWT 密钥，并在测试结束时还原。
// config.Config() 是惰性单例（返回 *AppConfig），jwtUtil 等直接读取它，因此覆写即生效。
func SetJWTSecret(t *testing.T, secret string) {
	t.Helper()
	cfg := config.Config()
	prev := cfg.Security.JWTSecret
	cfg.Security.JWTSecret = []byte(secret)
	t.Cleanup(func() { cfg.Security.JWTSecret = prev })
}
