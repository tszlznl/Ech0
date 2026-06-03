// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/lin-snow/ech0/internal/kvstore"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
)

// KV miss 时统一加载器报 AGENT_SETTING_NOT_FOUND，且不写默认行（best-effort 只读）。
// 用 kvstore.Memory（空）即可制造 miss，无需手搓替身。
func TestAgentSetting_MissReturnsNotFound(t *testing.T) {
	s := &CopilotService{durableKV: kvstore.NewMemory()}
	_, err := s.agentSetting(context.Background())
	if err == nil || err.Error() != commonModel.AGENT_SETTING_NOT_FOUND {
		t.Fatalf("expected AGENT_SETTING_NOT_FOUND, got %v", err)
	}
}

// KV 命中合法 JSON 时正确反序列化。
func TestAgentSetting_HappyPath(t *testing.T) {
	want := settingModel.AgentSetting{Enable: true, Protocol: "openai", Model: "gpt-x", ApiKey: "k"}
	raw, _ := json.Marshal(want)
	kv := kvstore.NewMemory()
	if err := kv.Set(context.Background(), commonModel.AgentSettingKey, string(raw)); err != nil {
		t.Fatalf("seed kv: %v", err)
	}
	s := &CopilotService{durableKV: kv}

	got, err := s.agentSetting(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Protocol != want.Protocol || got.Model != want.Model || !got.Enable {
		t.Fatalf("decoded setting mismatch: got %+v want %+v", got, want)
	}
}
