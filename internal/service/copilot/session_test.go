// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"strings"
	"testing"
	"time"

	"github.com/lin-snow/ech0/internal/agent"
	embeddingModel "github.com/lin-snow/ech0/internal/model/embedding"
)

func src(content string) embeddingModel.SearchResult {
	return embeddingModel.SearchResult{Content: content, EchoCreated: 0}
}

func TestHistoryForModel_Empty(t *testing.T) {
	if got := historyForModel(nil, "zh-CN", maxHistoryTokens, time.UTC); len(got) != 0 {
		t.Fatalf("expected empty history, got %d messages", len(got))
	}
}

// 旧轮 Sources 被丢弃；仅「最近一条带 Sources 的 assistant」把检索原文折进文本。
func TestHistoryForModel_DropsOldSourcesFoldsRecent(t *testing.T) {
	msgs := []ChatMessage{
		{Role: "user", Content: "q1"},
		{Role: "assistant", Content: "a1", Sources: []embeddingModel.SearchResult{src("OLD_ECHO")}},
		{Role: "user", Content: "q2"},
		{Role: "assistant", Content: "a2", Sources: []embeddingModel.SearchResult{src("RECENT_ECHO")}},
	}

	got := historyForModel(msgs, "zh-CN", maxHistoryTokens, time.UTC)
	if len(got) != 4 {
		t.Fatalf("expected 4 messages, got %d", len(got))
	}

	// 时间正序，角色映射正确。
	wantRoles := []agent.Role{agent.RoleUser, agent.RoleAssistant, agent.RoleUser, agent.RoleAssistant}
	for i, r := range wantRoles {
		if got[i].Role != r {
			t.Fatalf("msg %d: want role %q, got %q", i, r, got[i].Role)
		}
	}

	// 旧轮（a1）的 Sources 不应出现在任何文本里。
	a1 := got[1].Content
	if a1 != "a1" {
		t.Fatalf("old assistant content should stay plain, got %q", a1)
	}
	for _, m := range got {
		if strings.Contains(m.Content, "OLD_ECHO") {
			t.Fatalf("old sources should be dropped, but found in %q", m.Content)
		}
	}

	// 最近一轮（a2）应折入其检索原文。
	a2 := got[3].Content
	if !strings.Contains(a2, "a2") || !strings.Contains(a2, "RECENT_ECHO") {
		t.Fatalf("recent assistant should fold in its sources, got %q", a2)
	}
}

// Content 为空且无 Sources 的消息应被跳过。
func TestHistoryForModel_SkipsEmpty(t *testing.T) {
	msgs := []ChatMessage{
		{Role: "user", Content: "q1"},
		{Role: "assistant", Content: ""}, // 空且无 sources → 跳过
		{Role: "user", Content: "q2"},
		{Role: "assistant", Content: "a2"},
	}

	got := historyForModel(msgs, "en-US", maxHistoryTokens, time.UTC)
	if len(got) != 3 {
		t.Fatalf("expected 3 messages (empty skipped), got %d", len(got))
	}
	for _, m := range got {
		if strings.TrimSpace(m.Content) == "" {
			t.Fatalf("empty message should have been skipped")
		}
	}
}

// 超 token 预算时只保留最近若干条，且返回为时间正序。
func TestHistoryForModel_TokenBudgetKeepsRecentInOrder(t *testing.T) {
	// 每条 10 个 rune，预算 25 → 只能容纳最近 2 条。
	msgs := []ChatMessage{
		{Role: "user", Content: strings.Repeat("a", 10)},
		{Role: "assistant", Content: strings.Repeat("b", 10)},
		{Role: "user", Content: strings.Repeat("c", 10)},
		{Role: "assistant", Content: strings.Repeat("d", 10)},
	}

	got := historyForModel(msgs, "zh-CN", 25, time.UTC)
	if len(got) != 2 {
		t.Fatalf("expected 2 messages within budget, got %d", len(got))
	}
	// 时间正序：倒数第二条在前，最近一条在后。
	if got[0].Content != strings.Repeat("c", 10) || got[1].Content != strings.Repeat("d", 10) {
		t.Fatalf("expected most-recent two in time order, got [%q, %q]", got[0].Content, got[1].Content)
	}
}

// 极小预算下至少保留最近一条，不返回空、不 panic。
func TestHistoryForModel_TinyBudgetKeepsAtLeastOne(t *testing.T) {
	msgs := []ChatMessage{
		{Role: "user", Content: "q1"},
		{Role: "assistant", Content: strings.Repeat("z", 100)},
	}

	got := historyForModel(msgs, "zh-CN", 1, time.UTC)
	if len(got) != 1 {
		t.Fatalf("expected exactly 1 message under tiny budget, got %d", len(got))
	}
	if got[0].Content != strings.Repeat("z", 100) {
		t.Fatalf("expected the most-recent message to be kept, got %q", got[0].Content)
	}
}
