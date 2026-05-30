// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package agent

import (
	"strings"
	"testing"
)

// runGuard 把若干 chunk 依次喂进守卫，返回累计放行文本与是否触发泄漏。
func runGuard(chunks ...string) (emitted string, tripped bool) {
	g := &toolCallLeakGuard{}
	var b strings.Builder
	for _, c := range chunks {
		safe, hit := g.feed(c)
		b.WriteString(safe)
		if hit {
			return b.String(), true
		}
	}
	b.WriteString(g.flush())
	return b.String(), false
}

// 正常回答文本零损放行，不误触发。
func TestLeakGuard_NormalTextPassesThrough(t *testing.T) {
	in := []string{"根据你的 Echo，", "你最近在读《", "三体》，", "状态不错 🙂"}
	emitted, tripped := runGuard(in...)
	if tripped {
		t.Fatalf("normal text should not trip the guard")
	}
	if want := strings.Join(in, ""); emitted != want {
		t.Fatalf("text should pass through losslessly:\n want %q\n got  %q", want, emitted)
	}
}

// 单 chunk 内出现完整标记即触发。
func TestLeakGuard_FullMarkerInOneChunk(t *testing.T) {
	if _, tripped := runGuard("好的<tool_call> <function=search_echos>"); !tripped {
		t.Fatalf("a full <tool_call> marker should trip the guard")
	}
	if _, tripped := runGuard("x<function=search_echos>"); !tripped {
		t.Fatalf("a full <function= marker should trip the guard")
	}
}

// 标记被拆到两个 chunk 也要拼出来并触发（holdback 防漏）。
func TestLeakGuard_MarkerSplitAcrossChunks(t *testing.T) {
	emitted, tripped := runGuard("前文abc<tool", "_call>技术")
	if !tripped {
		t.Fatalf("a marker split across chunks should still trip")
	}
	// 触发前已放行的安全前缀不应包含任何标记片段。
	if strings.Contains(emitted, "<tool") {
		t.Fatalf("held partial marker must not leak before tripping, got %q", emitted)
	}
}

// 末尾「半个标记」但后续证明不是标记 → 暂留后照常放行，且零损。
func TestLeakGuard_PartialPrefixReleasedWhenNotMarker(t *testing.T) {
	in := []string{"完成<fun", "ny 想法"} // "<fun" 是 <function= 的前缀，但最终是 "<funny"
	emitted, tripped := runGuard(in...)
	if tripped {
		t.Fatalf("a partial prefix that resolves to non-marker should not trip")
	}
	if want := strings.Join(in, ""); emitted != want {
		t.Fatalf("released partial prefix should be lossless:\n want %q\n got  %q", want, emitted)
	}
}

// 含多字节中文且夹杂 '<' 的正常文本不应被错误截断或误触发。
func TestLeakGuard_MultibyteSafe(t *testing.T) {
	in := []string{"a < b 且 c < d，", "继续写想法"}
	emitted, tripped := runGuard(in...)
	if tripped {
		t.Fatalf("plain '<' in prose should not trip")
	}
	if want := strings.Join(in, ""); emitted != want {
		t.Fatalf("multibyte text should be lossless:\n want %q\n got  %q", want, emitted)
	}
}
