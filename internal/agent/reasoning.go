// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package agent

import (
	"regexp"
	"strings"
)

// thinkBlockRe 匹配模型把推理（reasoning）内联在正文里的 <think>...</think> 块。
// DeepSeek-R1 / QwQ 等推理模型经 OpenAI 兼容端返回时，常把思维过程塞进 content 而非独立的
// reasoning_content 字段；(?is) 开启大小写无关 + 让 . 跨行，非贪婪以正确处理多个块。
// 同时兼容 <thinking> 写法与可能带属性的开标签。
var thinkBlockRe = regexp.MustCompile(`(?is)<think(?:ing)?\b[^>]*>.*?</think(?:ing)?\s*>`)

// stripReasoning 去掉非流式生成结果里内联的推理标签块，只保留真正的回答。
// 这是模型/端点的输出怪癖（推理本不该混进答案），由 LLM 核心层统一收口；
// 上层（如近期总结 Widget）拿到的就是干净文本，无需各自再处理。
func stripReasoning(s string) string {
	return strings.TrimSpace(thinkBlockRe.ReplaceAllString(s, ""))
}

// thinkOpenMarkers / thinkCloseMarkers 是流式正文里推理块的起止标记（裸标签，不含属性——
// 推理模型内联输出的就是裸 <think>/<thinking>）。匹配大小写无关。
var (
	thinkOpenMarkers  = []string{"<think>", "<thinking>"}
	thinkCloseMarkers = []string{"</think>", "</thinking>"}
)

// reasoningSplitter 在流式正文（delta.content）里把内联的 <think>…</think> 推理段从答案里拆出来：
// 思考段归 reasoning，其余归 answer。标记可能被 chunk 边界切开，故末尾「可能是半个标记」的尾巴
// 暂留到下一片能判定为止（思路同 toolCallLeakGuard）。模型无关、只认裸标签。
//
// 注意：仅处理「reasoning 内联在 content」这一种来源；端点用独立 reasoning_content 字段时不会进这里。
type reasoningSplitter struct {
	inThink bool   // 当前是否处于 <think> 块内
	pending string // 尚不能判定归属的暂留尾巴
}

// feed 吃一段正文增量，返回本次可外放的答案文本与推理文本。
func (r *reasoningSplitter) feed(s string) (answer, reasoning string) {
	r.pending += s
	var ans, rea strings.Builder
	for {
		if r.inThink {
			idx, mlen := firstFoldMarker(r.pending, thinkCloseMarkers)
			if idx < 0 {
				hold := foldPrefixHold(r.pending, thinkCloseMarkers)
				rea.WriteString(r.pending[:len(r.pending)-hold])
				r.pending = r.pending[len(r.pending)-hold:]
				break
			}
			rea.WriteString(r.pending[:idx])
			r.pending = r.pending[idx+mlen:]
			r.inThink = false
			continue
		}
		idx, mlen := firstFoldMarker(r.pending, thinkOpenMarkers)
		if idx < 0 {
			hold := foldPrefixHold(r.pending, thinkOpenMarkers)
			ans.WriteString(r.pending[:len(r.pending)-hold])
			r.pending = r.pending[len(r.pending)-hold:]
			break
		}
		ans.WriteString(r.pending[:idx])
		r.pending = r.pending[idx+mlen:]
		r.inThink = true
	}
	return ans.String(), rea.String()
}

// flush 在流结束时放行暂留尾巴：仍在 think 块内（未闭合）归推理，否则归答案。
func (r *reasoningSplitter) flush() (answer, reasoning string) {
	rest := r.pending
	r.pending = ""
	if r.inThink {
		return "", rest
	}
	return rest, ""
}

// firstFoldMarker 返回 markers 中在 s 内最早出现的位置（大小写无关）及该标记长度；都没有则 idx=-1。
func firstFoldMarker(s string, markers []string) (idx, mlen int) {
	idx = -1
	for _, m := range markers {
		if i := foldIndex(s, m); i >= 0 && (idx < 0 || i < idx) {
			idx, mlen = i, len(m)
		}
	}
	return idx, mlen
}

// foldIndex 返回 marker（ASCII）在 s 中首次出现的下标，大小写无关；未找到返回 -1。
func foldIndex(s, marker string) int {
	for i := 0; i+len(marker) <= len(s); i++ {
		if foldHasPrefix(s[i:], marker) {
			return i
		}
	}
	return -1
}

// foldPrefixHold 返回 s 末尾「可能是某个 marker 前缀」的最长长度——这段需暂留待后续 chunk 判定。
func foldPrefixHold(s string, markers []string) int {
	hold := 0
	for _, m := range markers {
		n := min(len(m), len(s))
		for k := n; k > hold; k-- {
			if foldHasPrefix(m, s[len(s)-k:]) {
				hold = k
				break
			}
		}
	}
	return hold
}

// foldHasPrefix 报告 s 是否以 prefix 开头（按 ASCII 大小写无关）。
func foldHasPrefix(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	for i := 0; i < len(prefix); i++ {
		if asciiLower(s[i]) != asciiLower(prefix[i]) {
			return false
		}
	}
	return true
}

// asciiLower 把 ASCII 大写字母转小写，其余字节原样返回。
func asciiLower(b byte) byte {
	if 'A' <= b && b <= 'Z' {
		return b + ('a' - 'A')
	}
	return b
}
