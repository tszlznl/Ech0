// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
)

// fakeKV 是 KeyValueRepository 的测试替身：get 返回固定 value/err，写操作记录最后一次。
type fakeKV struct {
	value string
	err   error
}

func (f *fakeKV) GetKeyValue(_ context.Context, _ string) (string, error) {
	return f.value, f.err
}
func (f *fakeKV) AddOrUpdateKeyValue(_ context.Context, _, _ string) error { return nil }
func (f *fakeKV) DeleteKeyValue(_ context.Context, _ string) error         { return nil }

// KV miss 时统一加载器报 AGENT_SETTING_NOT_FOUND，且不写默认行（best-effort 只读）。
func TestAgentSetting_MissReturnsNotFound(t *testing.T) {
	s := &CopilotService{kvRepository: &fakeKV{err: errors.New("not found")}}
	_, err := s.agentSetting(context.Background())
	if err == nil || err.Error() != commonModel.AGENT_SETTING_NOT_FOUND {
		t.Fatalf("expected AGENT_SETTING_NOT_FOUND, got %v", err)
	}
}

// KV 命中合法 JSON 时正确反序列化。
func TestAgentSetting_HappyPath(t *testing.T) {
	want := settingModel.AgentSetting{Enable: true, Protocol: "openai", Model: "gpt-x", ApiKey: "k"}
	raw, _ := json.Marshal(want)
	s := &CopilotService{kvRepository: &fakeKV{value: string(raw)}}

	got, err := s.agentSetting(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Protocol != want.Protocol || got.Model != want.Model || !got.Enable {
		t.Fatalf("decoded setting mismatch: got %+v want %+v", got, want)
	}
}
