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

// localeIsZH 判定是否中文 locale（zh-* 走中文，其余回退英文）。
func localeIsZH(locale string) bool {
	return strings.HasPrefix(strings.ToLower(locale), "zh")
}

// runStringsFor 按 locale 选择 Loop 回喂/注入模型的提示文案（agent.RunStrings）。
// 让回喂模型的文本随用户语言切换，而非固定中文。
func runStringsFor(locale string) agent.RunStrings {
	if localeIsZH(locale) {
		return agent.RunStrings{
			DedupNote:   "（已检索过，结果见上）",
			UnknownTool: "未知工具：",
			ToolError:   "工具执行失败：",
			ImageNote:   "（以下是上一步检索命中的 Echo 的配图，供你结合图片内容作答）",
		}
	}
	return agent.RunStrings{
		DedupNote:   "(Already searched; see the results above.)",
		UnknownTool: "Unknown tool: ",
		ToolError:   "Tool execution failed: ",
		ImageNote:   "(Below are images from the Echo matched in the previous step; use them to inform your answer.)",
	}
}

// chatSystemPrompt 是 Chat（Agent 形态）的系统提示词：声明工具用途与作答纪律。
const chatSystemPrompt = `你是用户自己的 Echo（微博客/碎碎念）回顾助手。
你有两个工具，按需选用：
- search_echos：点查。回答具体问题、找某几条相关记录时用它（top-k，只返回最相关的若干条，是采样不是全貌）。
- summarize_echos：区间聚合。当用户要「某段时间的总结/回顾」（年终、年度、季度、月度，或“上半年发了什么”这类）时用它——它会覆盖该区间内的【全部】Echo。
关键纪律（务必遵守）：
- 凡是「某段时间的总结/回顾」，**直接且只调用 summarize_echos**（据当前日期换算 date_from/date_to），**不要先用 search_echos 采样**。summarize_echos 返回的材料才是完整依据。
- 写这类总结时，**严格依据 summarize_echos 的聚合材料**，覆盖材料里的各个月份/各条主线，不要只挑某几条生动的展开、不要把少量样本当成全貌。材料里的 #标签、[img×N]（配图数）、[音乐/网站/位置…] 等都是线索，可用于归纳主题与活跃度。
- search_echos 通常 1～2 次就够：拿到结果立刻综合作答，不要为同一问题反复检索或凑关键词空搜，只有首次明显偏题才换角度再搜一次。
作答要求：
- 优先依据工具返回的内容作答，做跨条目/跨时间的归纳、总结与回顾；
- 如果没有足够依据，就如实说明“你的 Echo 里没有相关记录”，不要编造；若材料标注了覆盖范围或截断，请在总结中如实体现；
- 用简洁自然的中文，可用 Emoji 和换行，不要输出 HTML 标签。`

// chatSystemPromptEN 是 chatSystemPrompt 的英文版本（locale 非 zh-* 时使用）。
const chatSystemPromptEN = `You are the user's personal assistant for reviewing their Echos (microblog notes).
You have two tools; pick the right one:
- search_echos: pinpoint lookup. Use it to answer specific questions or find a few relevant entries (top-k, returns only the most relevant ones).
- summarize_echos: range aggregation. Use it when the user wants a "summary/review of a time period" (year-end, yearly, quarterly, monthly, etc.) — it covers ALL Echos in that range, not just a sample. Always use it for year-end/annual summaries, converting the current date into date_from/date_to.
Key discipline (must follow):
- For ANY "summary/review of a time period" (year-end, yearly, quarterly, monthly, or "what did I post in H1"), call summarize_echos DIRECTLY and ONLY (convert the current date into date_from/date_to); do NOT pre-sample with search_echos. Its returned material is the complete basis.
- When writing such a summary, ground it STRICTLY in the summarize_echos material, covering the various months / main threads in it; do not just expand a few vivid entries and do not treat a small sample as the whole. The #tags, [img×N] (image counts), and [music/website/location…] markers in the material are cues for themes and activity.
- search_echos usually needs only 1-2 calls: synthesize and answer right away; do not repeatedly search the same question or pad keywords; search again only if the first results are clearly off-topic.
Answering requirements:
- Prefer answering based on the tool's returned content, doing cross-entry / cross-time synthesis, summary and review;
- If there is not enough evidence, honestly say "there are no relevant records in your Echos"; do not make things up; if the material notes its coverage or that it was truncated, reflect that honestly in your summary;
- Be concise and natural, you may use emoji and line breaks, do not output HTML tags.
Always answer in the same language as the user's question.`

// buildChatMessages 组装 Chat 一轮对话的消息：system → 历史多轮 → 本轮问题。
// today（YYYY-MM-DD）与 tagNames 作为检索上下文拼进 system，让模型能把相对时间换算成
// date_from/date_to、并从已知标签里挑 tags。
// history 是经 historyForModel 裁剪好的过往多轮（已剥旧工具结果、按 token 预算截断）。
func buildChatMessages(history []agent.Message, question, locale, today string, tagNames []string) []agent.Message {
	msgs := make([]agent.Message, 0, len(history)+2)
	msgs = append(msgs, agent.Message{Role: agent.RoleSystem, Content: buildSystemPrompt(locale, today, tagNames)})
	msgs = append(msgs, history...)
	msgs = append(msgs, agent.Message{Role: agent.RoleUser, Content: question})
	return msgs
}

// buildSystemPrompt 拼出 Chat 的完整 system 提示词（系统纪律 + 检索上下文块）。
// 抽成独立函数，供 AskStream 在裁剪历史前估算固定开销 token。
func buildSystemPrompt(locale, today string, tagNames []string) string {
	return chatSystemPromptFor(locale) + buildContextBlock(locale, today, tagNames)
}

// buildContextBlock 生成注入 system prompt 的检索上下文块（当前日期 + 可用标签）。
func buildContextBlock(locale, today string, tagNames []string) string {
	var b strings.Builder
	if localeIsZH(locale) {
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
	if localeIsZH(locale) {
		return chatSystemPrompt
	}
	return chatSystemPromptEN
}

// recentSourcesNote{ZH,EN} 是折进「最近一轮 assistant 文本」的检索依据标注块（含一个 %s 占位
// 给 formatSearchResults 的命中文本），让用户追问上一轮结果细节时模型直接有据，无需重检索。
const recentSourcesNoteZH = "（上一轮检索到的 Echo 依据，供追问其细节时参考：\n%s）"

const recentSourcesNoteEN = "(Echos retrieved in the previous turn, for reference when asked about their details:\n%s)"

// recentSourcesNoteFor 按 locale 选择检索依据标注块文案。
func recentSourcesNoteFor(locale string) string {
	if localeIsZH(locale) {
		return recentSourcesNoteZH
	}
	return recentSourcesNoteEN
}

// summarySystemPromptFor 按 locale 选择「近况总结」系统提示词。
func summarySystemPromptFor(locale string) string {
	if localeIsZH(locale) {
		return summarySystemPromptZH
	}
	return summarySystemPromptEN
}

// summaryUserPromptFor 按 locale 选择「近况总结」用户提示词。
func summaryUserPromptFor(locale string) string {
	if localeIsZH(locale) {
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

// aggregateMapPromptFor 是区间聚合 map 阶段（单月浓缩）的指令：把一个月的 Echo 压成
// 事实性摘要供上层再归纳，强调忠实、保留关键信息、不发挥、随内容语言。
func aggregateMapPromptFor(locale string) string {
	if localeIsZH(locale) {
		return "下面是用户在一段时间内发布的若干 Echo（按月分组，每条含日期，可能带 #标签、[img×N]（配图数）、" +
			"[音乐/网站/位置…] 等线索）。请把它们浓缩成一段紧凑、忠实的事实性摘要，保留关键事件、反复出现的主题、" +
			"心情变化、提到的人/地点/作品、活跃的标签与发图情况；只做归纳，不要发挥或编造，不要逐条罗列、不要输出 HTML。" +
			"用与内容相同的语言。这是给后续年度/区间总结使用的中间材料。"
	}
	return "Below are several Echos the user posted over a period (grouped by month, each with a date and possibly " +
		"#tags, [img×N] (image count), [music/website/location…] cues). Condense them into a compact, faithful factual digest that " +
		"preserves key events, recurring themes, mood shifts, mentioned people/places/works, active tags and posting of images. " +
		"Summarize only — do not embellish or invent, do not enumerate each entry, and do not output HTML. " +
		"Use the same language as the content. This is intermediate material for a later period summary."
}

// searchCoverageNoteFor 在 search_echos 命中数多于本次展示（top-k 截断）时，给模型的一行如实告知，
// 防止它把「采样的几条」当成「全部」。
func searchCoverageNoteFor(locale string, total, shown int) string {
	if localeIsZH(locale) {
		return fmt.Sprintf("（本次条件共命中 %d 条，下面只展示最相关的 %d 条；若需覆盖全部用于总结/回顾，请改用 summarize_echos。）", total, shown)
	}
	return fmt.Sprintf("(This filter matched %d Echos in total; only the %d most relevant are shown below. To cover all of them for a summary/review, use summarize_echos instead.)", total, shown)
}

// aggregateMaterialHeaderFor 是 summarize_echos 回喂模型的物料抬头：声明这是覆盖某区间的中间材料、
// 覆盖度多少、是否截断，并提示模型据此为用户写最终成稿（而非把材料原样吐回）。
func aggregateMaterialHeaderFor(locale string, total, returned, buckets int, truncated bool) string {
	if localeIsZH(locale) {
		var b strings.Builder
		fmt.Fprintf(&b, "以下是该时间区间内 Echo 的聚合材料（共命中 %d 条，已纳入 %d 条", total, returned)
		if buckets > 1 {
			fmt.Fprintf(&b, "，因体量较大已按月分层浓缩为 %d 段", buckets)
		}
		b.WriteString("）。")
		if truncated {
			fmt.Fprintf(&b, "注意：区间内条数超过单次上限，已保留最近的 %d 条，请在总结中说明这一点。", returned)
		}
		b.WriteString("请据此为用户撰写最终的总结/回顾，做跨时间的归纳，不要逐条复述材料。")
		return b.String()
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Below is aggregated material for the Echos in this time range (%d matched in total, %d included", total, returned)
	if buckets > 1 {
		fmt.Fprintf(&b, "; due to volume it was condensed month-by-month into %d sections", buckets)
	}
	b.WriteString("). ")
	if truncated {
		fmt.Fprintf(&b, "Note: the range exceeded the per-run cap, so only the most recent %d were kept — mention this in your summary. ", returned)
	}
	b.WriteString("Use it to write the final summary/review for the user, synthesizing across time rather than restating each entry.")
	return b.String()
}

// aggregateReducePromptFor 是 reduce 阶段的指令：当各月摘要拼接后仍超预算时，再压一轮，
// 但必须保留每个月的要点与时间线，不丢月份。
func aggregateReducePromptFor(locale string) string {
	if localeIsZH(locale) {
		return "下面是按月排好的多段摘要。请进一步压缩成更短的分月要点，保留每个月的核心信息与时间线，" +
			"不要遗漏任何月份，不要发挥或编造，不要输出 HTML。用与内容相同的语言。这是给后续年度/区间总结使用的中间材料。"
	}
	return "Below are several month-by-month digests in order. Compress them further into shorter per-month key " +
		"points, preserving each month's core information and the timeline. Do not drop any month, do not " +
		"embellish or invent, and do not output HTML. Use the same language as the content. This is intermediate material for a later period summary."
}
