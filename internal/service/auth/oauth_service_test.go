// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package auth

import (
	"testing"

	"github.com/lin-snow/ech0/internal/config"
)

func TestParseAndValidateClientRedirect_Allowed(t *testing.T) {
	cfg := config.Config()
	cfg.Auth.Redirect.AllowedReturnURLs = []string{"https://app.example.com/auth"}

	svc := &AuthService{}
	u, err := svc.parseAndValidateClientRedirect("https://app.example.com/auth?from=test")
	if err != nil {
		t.Fatalf("expected allow redirect, got err: %v", err)
	}
	if u.Host != "app.example.com" {
		t.Fatalf("unexpected host: %s", u.Host)
	}
}

func TestParseAndValidateClientRedirect_Denied(t *testing.T) {
	cfg := config.Config()
	cfg.Auth.Redirect.AllowedReturnURLs = []string{"https://app.example.com/auth"}

	svc := &AuthService{}
	_, err := svc.parseAndValidateClientRedirect("https://evil.example.net/auth")
	if err == nil {
		t.Fatalf("expected deny redirect")
	}
}

// 防 GHSA-p64j-f4x9-wq66：scheme+host 相同但 path 不在白名单内必须拒绝，否则
// 攻击者可借助同源任意路径上的 Referer/分析脚本/open-redirect 截获一次性
// exchange code。
func TestParseAndValidateClientRedirect_PathMismatchDenied(t *testing.T) {
	cfg := config.Config()
	cfg.Auth.Redirect.AllowedReturnURLs = []string{"https://app.example.com/auth"}

	svc := &AuthService{}
	_, err := svc.parseAndValidateClientRedirect(
		"https://app.example.com/attacker-chosen-path",
	)
	if err == nil {
		t.Fatalf("expected deny redirect with mismatched path")
	}
}
