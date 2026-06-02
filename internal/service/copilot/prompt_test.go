// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"strings"
	"testing"
)

// buildContextBlock 须注入一行**任务中性**的身份信息：交代「跟谁对话 + 检索到的是谁的 Echo」，
// 但不带「回顾」等具体任务动词（Chat 不一定是回顾，也可能是查找/延伸/找灵感）。
func TestBuildContextBlock_IdentityLine(t *testing.T) {
	zh := buildContextBlock("zh-CN", "2026-06-02", nil, "Alice")
	if !strings.Contains(zh, "Alice") {
		t.Fatalf("zh context block should mention the display name: %q", zh)
	}
	if !strings.Contains(zh, "本人发布") {
		t.Fatalf("zh context block should scope retrieval to the user: %q", zh)
	}
	if strings.Contains(zh, "回顾") {
		t.Fatalf("identity line must be task-neutral (no 回顾): %q", zh)
	}

	en := buildContextBlock("en", "2026-06-02", nil, "Alice")
	if !strings.Contains(en, "Alice") || !strings.Contains(en, "posted themselves") {
		t.Fatalf("en context block should mention name + scope: %q", en)
	}

	// 空展示名 → 省略身份行（仍保留日期块），不输出空名。
	if got := buildContextBlock("zh-CN", "2026-06-02", nil, ""); strings.Contains(got, "当前与你对话的是") {
		t.Fatalf("empty display name should omit identity line: %q", got)
	}
}

// 系统提示词应是通用「私人助手」定位，而非写死的「回顾助手」，但工具仍齐备。
func TestChatSystemPrompt_GeneralFraming(t *testing.T) {
	zh := chatSystemPromptFor("zh-CN")
	if strings.Contains(zh, "回顾助手") {
		t.Fatalf("system prompt should not be framed as a review-only assistant: %q", zh)
	}
	if !strings.Contains(zh, "私人助手") {
		t.Fatalf("system prompt should be framed as a general personal assistant: %q", zh)
	}
	for _, tool := range []string{"search_echos", "summarize_echos", "stats_overview"} {
		if !strings.Contains(zh, tool) {
			t.Fatalf("system prompt should still declare tool %q: %q", tool, zh)
		}
	}

	en := chatSystemPromptFor("en")
	if !strings.Contains(en, "personal assistant") {
		t.Fatalf("en system prompt should be framed as a personal assistant: %q", en)
	}
}
