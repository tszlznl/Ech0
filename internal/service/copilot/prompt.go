// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"fmt"
	"sort"
	"strings"

	"github.com/lin-snow/ech0/internal/agent"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
)

// maxPromptTags 限制注入 system prompt 的标签数量，避免标签很多时占用过多 token。
const maxPromptTags = 40

// chatSystemPrompt 是 Chat（Agent 形态）的系统提示词：声明工具用途与作答纪律。
const chatSystemPrompt = `你是用户自己的 Echo（微博客/碎碎念）回顾助手。
你可以调用 search_echos 工具检索用户过往发布的 Echo。
检索策略（重要）：
- 当问题涉及用户的历史记录、需要具体依据时，先检索一次；通常检索 1～2 次就足够；
- 拿到检索结果后，请立刻综合这些结果直接作答，不要为同一问题反复检索、不要凑关键词空搜；
- 只有当首次结果明显偏题、确需换一个角度时才再检索一次。
作答要求：
- 优先依据检索到的 Echo 内容作答，做跨条目的归纳、总结与回顾；
- 如果检索不到足够依据，就如实说明“你的 Echo 里没有相关记录”，不要编造；
- 用简洁自然的中文，可用 Emoji 和换行，不要输出 HTML 标签。`

// chatSystemPromptEN 是 chatSystemPrompt 的英文版本（locale 非 zh-* 时使用）。
const chatSystemPromptEN = `You are the user's personal assistant for reviewing their Echos (microblog notes).
You can call the search_echos tool to retrieve the user's previously published Echos.
Retrieval strategy (important):
- When the question concerns the user's history or needs concrete evidence, search once first; usually 1-2 searches are enough;
- Once you have results, synthesize them and answer directly right away; do not repeatedly search for the same question, and do not run empty searches by padding keywords;
- Only search again with a different angle if the first results are clearly off-topic.
Answering requirements:
- Prefer answering based on the retrieved Echo content, doing cross-entry synthesis, summary and review;
- If you cannot find enough evidence, honestly say "there are no relevant records in your Echos"; do not make things up;
- Be concise and natural, you may use emoji and line breaks, do not output HTML tags.
Always answer in the same language as the user's question.`

// buildChatMessages 组装 Chat 一轮对话的初始消息（system + 用户问题）。
// today（YYYY-MM-DD）与 tagNames 作为检索上下文拼进 system，让模型能把相对时间换算成
// date_from/date_to、并从已知标签里挑 tags。
// 历史多轮接入点：未来在此把会话历史拼进 system 与 question 之间即可（设计 §11）。
func buildChatMessages(question, locale, today string, tagNames []string) []agent.Message {
	sys := chatSystemPromptFor(locale) + buildContextBlock(locale, today, tagNames)
	return []agent.Message{
		{Role: agent.RoleSystem, Content: sys},
		{Role: agent.RoleUser, Content: question},
	}
}

// buildContextBlock 生成注入 system prompt 的检索上下文块（当前日期 + 可用标签）。
func buildContextBlock(locale, today string, tagNames []string) string {
	var b strings.Builder
	if strings.HasPrefix(strings.ToLower(locale), "zh") {
		fmt.Fprintf(&b, "\n\n当前日期：%s（涉及“去年/上个月/最近”等相对时间时，据此换算成 date_from/date_to 传给 search_echos）。", today)
		if len(tagNames) > 0 {
			fmt.Fprintf(&b, "\n用户可用标签：%s。需按标签筛选时，从中选取标签名传给 tags 参数。", strings.Join(tagNames, "、"))
		}
	} else {
		fmt.Fprintf(&b, "\n\nCurrent date: %s (use it to convert relative times like \"last year/last month/recently\" into date_from/date_to for search_echos).", today)
		if len(tagNames) > 0 {
			fmt.Fprintf(&b, "\nAvailable tags: %s. When filtering by tag, pick names from these for the tags argument.", strings.Join(tagNames, ", "))
		}
	}
	return b.String()
}

// tagNamesForPrompt 取使用次数最高的若干标签名注入 prompt（按 UsageCount 降序、上限
// maxPromptTags），在召回质量与 token 成本间取平衡。
func tagNamesForPrompt(tags []echoModel.Tag) []string {
	sorted := append([]echoModel.Tag(nil), tags...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].UsageCount > sorted[j].UsageCount })
	if len(sorted) > maxPromptTags {
		sorted = sorted[:maxPromptTags]
	}
	names := make([]string, 0, len(sorted))
	for _, t := range sorted {
		names = append(names, t.Name)
	}
	return names
}

// chatSystemPromptFor 按 locale 选择系统提示词：zh-* 用中文，其余统一回退英文。
func chatSystemPromptFor(locale string) string {
	if strings.HasPrefix(strings.ToLower(locale), "zh") {
		return chatSystemPrompt
	}
	return chatSystemPromptEN
}

// summarySystemPromptFor 按 locale 选择「近况总结」系统提示词。
func summarySystemPromptFor(locale string) string {
	if strings.HasPrefix(strings.ToLower(locale), "zh") {
		return summarySystemPromptZH
	}
	return summarySystemPromptEN
}

// summaryUserPromptFor 按 locale 选择「近况总结」用户提示词。
func summaryUserPromptFor(locale string) string {
	if strings.HasPrefix(strings.ToLower(locale), "zh") {
		return summaryUserPromptZH
	}
	return summaryUserPromptEN
}

const summarySystemPromptZH = `
					这是“近况总结”场景，请使用简洁自然的中文表达。
					不使用复杂格式：不要标题、列表、表格、代码块、链接。
					不要输出任何原始 HTML 标签。
					可使用纯文字、Emoji 和正常换行来增强可读性。
					回复保持简洁，聚焦作者最近的活动和状态。`

const summarySystemPromptEN = `
					This is a "recent status summary" scenario. Use concise and natural language.
					Do not use complex formatting: no headings, lists, tables, code blocks, or links.
					Do not output any raw HTML tags.
					You may use plain text, emoji and normal line breaks to improve readability.
					Keep the reply concise, focusing on the author's recent activity and state.`

const summaryUserPromptZH = "请根据提供的近期互动内容（内容可能包括日常生活、句子诗词摘抄、吐槽等等），总结该用户最近的活动和状态，突出作者状态即可，不需要详细描述内容，如果没有任何内容，请回复作者最近很神秘~"

const summaryUserPromptEN = "Based on the provided recent activity (which may include daily life, quoted sentences or poems, venting, etc.), summarize this user's recent activity and state. Just highlight the author's state without describing the content in detail. If there is no content at all, reply that the author has been quite mysterious lately~"
