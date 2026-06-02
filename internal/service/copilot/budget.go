// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import settingModel "github.com/lin-snow/ech0/internal/model/setting"

const (
	// defaultContextWindow 是未配置 AgentSetting.ContextWindow 时的保守假设（token）。
	// 取 256k：当前主流长上下文模型的常见下限，足以一次塞入一名普通用户一整年的 Echo。
	defaultContextWindow = 256_000
	// contextUsableRatio 是窗口中可用于「塞检索物料」的比例，其余留给 system prompt、
	// 工具定义、对话历史与模型生成的成稿。
	contextUsableRatio = 0.6
	// contextReserveTokens 是在比例之外再扣掉的固定开销冗余（system + tool def + 成稿留白）。
	contextReserveTokens = 8_000
	// minAggregateBudget 是预算下限：即便窗口配得极小，也保证每轮 map 有一个可用的压缩目标。
	minAggregateBudget = 2_000
)

// aggregateBudgetTokens 据模型窗口推算「区间聚合可塞入的 token 预算」。
// ContextWindow=0（未配置）按 defaultContextWindow 处理；结果不低于 minAggregateBudget。
// 复用 estimateTokens（rune 计数）做度量，无需引入 tokenizer。
func aggregateBudgetTokens(setting settingModel.AgentSetting) int {
	window := setting.ContextWindow
	if window <= 0 {
		window = defaultContextWindow
	}
	budget := int(float64(window)*contextUsableRatio) - contextReserveTokens
	if budget < minAggregateBudget {
		return minAggregateBudget
	}
	return budget
}

// chatContextBudgetTokens 是 Chat 工具循环里「整轮消息上下文」的软上限：超过即由 agent loop
// 回收最旧的工具结果（占位替换），防多轮工具结果累积撑爆窗口。复用与区间聚合相同的窗口预算口径。
func chatContextBudgetTokens(setting settingModel.AgentSetting) int {
	return aggregateBudgetTokens(setting)
}
