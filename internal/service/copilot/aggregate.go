// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lin-snow/ech0/internal/agent"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	embeddingModel "github.com/lin-snow/ech0/internal/model/embedding"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
)

// summarizeArgs 是 summarize_echos 的入参：date_from/date_to 必填（区间），tags/focus 可选。
type summarizeArgs struct {
	DateFrom string   `json:"date_from"`
	DateTo   string   `json:"date_to"`
	Tags     []string `json:"tags"`
	Focus    string   `json:"focus"`
}

// summarizeEchosTool 是注入给 agent 的领域工具：穷举聚合某时间区间内的【全部】 Echo，
// 供模型撰写跨度较长的总结/回顾（年终/年度/季度/月度）。与 search_echos 的 top-k 互补——
// 它覆盖区间全部而非采样，并随窗口预算自适应（放得下整段塞入，放不下按月 map-reduce）。
func (s *CopilotService) summarizeEchosTool(
	allTags []echoModel.Tag,
	setting settingModel.AgentSetting,
	locale string,
) agent.Tool {
	return agent.Tool{
		Def: agent.ToolDef{
			Name:        "summarize_echos",
			Description: "聚合某时间区间内的【全部】Echo，用于生成跨度较长的总结/回顾（如年终、年度、季度、月度总结）。它会覆盖区间内所有记录而非只采样几条，正是「帮我写年终/年度总结」这类需求该用的工具。必须提供 date_from 与 date_to；可选 tags（按标签名限定主题）、focus（侧重点，如“工作”“读书”“心情”）。",
			Parameters:  json.RawMessage(`{"type":"object","properties":{"date_from":{"type":"string","description":"起始日期，格式 YYYY-MM-DD，含当天"},"date_to":{"type":"string","description":"结束日期，格式 YYYY-MM-DD，含当天"},"tags":{"type":"array","items":{"type":"string"},"description":"可选，按标签名限定主题；可用标签见系统提示"},"focus":{"type":"string","description":"可选，总结的侧重点，如“工作”“读书”“心情”"}},"required":["date_from","date_to"]}`),
		},
		Execute: func(ctx context.Context, args json.RawMessage) (agent.ToolOutput, error) {
			var a summarizeArgs
			_ = json.Unmarshal(args, &a)
			from := parseDay(a.DateFrom, false)
			to := parseDay(a.DateTo, true)
			if from == 0 && to == 0 {
				return agent.ToolOutput{}, errors.New("summarize_echos 需要 date_from 与 date_to 指定时间区间")
			}
			tagIDs := resolveTagIDs(allTags, a.Tags)

			results, total, truncated, err := s.collectRange(ctx, tagIDs, from, to)
			if err != nil {
				return agent.ToolOutput{}, err
			}
			reverseResults(results) // 倒序拉取后翻成时间正序，便于按月归纳与顺读成稿

			material, buckets, err := s.mapReduceSummary(ctx, setting, locale, results, aggregateBudgetTokens(setting))
			if err != nil {
				return agent.ToolOutput{}, err
			}

			cov := aggregateCoverage{
				Total:     total,
				Returned:  len(results),
				Buckets:   buckets,
				Truncated: truncated,
			}
			header := aggregateMaterialHeaderFor(locale, int(total), len(results), buckets, truncated)
			return agent.ToolOutput{
				Content: header + "\n\n" + material,
				Meta:    cov,
			}, nil
		},
	}
}

const (
	// aggregatePageSize 是区间取全时每页拉取的条数（批量、减少往返）。
	aggregatePageSize = 200
	// maxAggregateEchos 是单次聚合纳入的硬上限：兜底防超大账号把一整年几万条全拉进内存/上下文。
	// 触顶按时间倒序保留「最近 N 条」，并在结果中如实标注截断，绝不静默丢弃。
	maxAggregateEchos = 5000
)

// aggregateCoverage 是 summarize_echos 的旁路覆盖度元数据（经 ToolOutput.Meta → SSE coverage 事件），
// 让模型与用户都清楚「这次总结覆盖了多少」，杜绝静默截断式的「看着很完整」。
type aggregateCoverage struct {
	Total     int64 `json:"total"`     // 区间内命中的 Echo 总数
	Returned  int   `json:"returned"`  // 实际纳入聚合的条数（受 maxAggregateEchos 约束）
	Buckets   int   `json:"buckets"`   // map-reduce 分桶数；1 表示窗口放得下、直接整段塞入
	Truncated bool  `json:"truncated"` // 是否因硬上限截断（保留最近）
}

// collectRange 穷举拉取 [from,to]（可叠加 tagIDs）区间内的全部 Echo，与 search_echos 的
// top-k 不同——它要的是「覆盖全部」而非「最相关几条」。按 created_at 倒序（QueryEchos 默认）
// 分页累加，触顶 maxAggregateEchos 即停并标记 truncated（因倒序，保留的是最近的）。
func (s *CopilotService) collectRange(
	ctx context.Context,
	tagIDs []string,
	from, to int64,
) (results []embeddingModel.SearchResult, total int64, truncated bool, err error) {
	for page := 1; ; page++ {
		res, qErr := s.echoService.QueryEchos(ctx, commonModel.EchoQueryDto{
			Page:     page,
			PageSize: aggregatePageSize,
			TagIDs:   tagIDs,
			DateFrom: from,
			DateTo:   to,
		})
		if qErr != nil {
			return nil, 0, false, qErr
		}
		total = res.Total
		for i := range res.Items {
			results = append(results, echoToSearchResult(res.Items[i]))
			if len(results) >= maxAggregateEchos {
				return results, total, total > int64(len(results)), nil
			}
		}
		// 本页未满或已覆盖总数 → 取尽。
		if len(res.Items) < aggregatePageSize || int64(len(results)) >= total {
			return results, total, false, nil
		}
	}
}

// echoToSearchResult 把一条 Echo 映射成检索结果形状，使聚合路径与检索路径同构，
// 复用 formatSearchResults / SSE sources 等下游逻辑。
func echoToSearchResult(e echoModel.Echo) embeddingModel.SearchResult {
	return embeddingModel.SearchResult{
		EchoID:      e.ID,
		Content:     e.Content,
		Username:    e.Username,
		EchoCreated: e.CreatedAt,
		Distance:    0,
	}
}

// reverseResults 把倒序（最新在前）翻成时间正序（最早在前），便于按月分桶与顺读成稿。
func reverseResults(results []embeddingModel.SearchResult) {
	for l, r := 0, len(results)-1; l < r; l, r = l+1, r-1 {
		results[l], results[r] = results[r], results[l]
	}
}

// mapReduceSummary 据窗口预算把区间内的 Echo 浓缩成「供模型写最终成稿」的中间材料：
//   - 放得下（estimateTokens ≤ budget）→ 直接返回完整格式化文本（buckets=1），大窗口在此吃满；
//   - 放不下 → 按月分桶，每桶 LLM map 成事实性摘要，拼接；若拼接仍超预算，再 reduce 一轮。
//
// map 阶段 v1 顺序执行（成本可预测、实现简单）；并行（bounded）留作后续优化。
// 任一 map/reduce 调用失败即上抛，由 agent loop 作为「工具执行失败」回喂模型自愈。
func (s *CopilotService) mapReduceSummary(
	ctx context.Context,
	setting settingModel.AgentSetting,
	locale string,
	results []embeddingModel.SearchResult,
	budget int,
) (content string, buckets int, err error) {
	full := formatSearchResults(results, nil)
	if estimateTokens(full) <= budget {
		return full, 1, nil
	}

	// 超预算：按 YYYY-MM 分桶（保持时间正序）。
	type bucket struct {
		month string
		items []embeddingModel.SearchResult
	}
	var ordered []bucket
	index := make(map[string]int)
	for _, r := range results {
		m := time.Unix(r.EchoCreated, 0).UTC().Format("2006-01")
		if i, ok := index[m]; ok {
			ordered[i].items = append(ordered[i].items, r)
			continue
		}
		index[m] = len(ordered)
		ordered = append(ordered, bucket{month: m, items: []embeddingModel.SearchResult{r}})
	}

	var b strings.Builder
	for _, bk := range ordered {
		digest, mErr := agent.Generate(ctx, setting, []agent.Message{
			{Role: agent.RoleSystem, Content: aggregateMapPromptFor(locale)},
			{Role: agent.RoleUser, Content: formatSearchResults(bk.items, nil)},
		}, false, nil)
		if mErr != nil {
			return "", 0, mErr
		}
		fmt.Fprintf(&b, "【%s】\n%s\n\n", bk.month, strings.TrimSpace(digest))
	}
	joined := strings.TrimSpace(b.String())

	// 各月摘要拼接后仍超预算 → 再压一轮（保留每月要点与时间线）。
	if estimateTokens(joined) > budget {
		reduced, rErr := agent.Generate(ctx, setting, []agent.Message{
			{Role: agent.RoleSystem, Content: aggregateReducePromptFor(locale)},
			{Role: agent.RoleUser, Content: joined},
		}, false, nil)
		if rErr != nil {
			return "", 0, rErr
		}
		joined = strings.TrimSpace(reduced)
	}

	return joined, len(ordered), nil
}
