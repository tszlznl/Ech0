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
)

// 检索默认返回的命中条数（固定，不暴露给模型，行为可预测）
const defaultTopK = 6

// searchArgs 是 search_echos 工具的入参。三者皆可选，但至少需其一：
// query 走语义/关键词，tags/date_* 走结构化精确过滤。
type searchArgs struct {
	Query    string   `json:"query"`
	Tags     []string `json:"tags"`
	DateFrom string   `json:"date_from"`
	DateTo   string   `json:"date_to"`
}

// searchEchosTool 是注入给 agent 的领域工具：检索用户过往 Echo。
//
// 检索路由（一条规则）：只要带结构化过滤（tags / date_*）就走 QueryEchos（SQL 精确，
// 向量索引做不了元数据过滤）；纯 query 且向量已启用才走 embedding 语义检索，否则回退
// QueryEchos 的 content LIKE。allTags 用于把模型给的标签名解析成 ID（UUID 不进 prompt）。
func (s *CopilotService) searchEchosTool(allTags []echoModel.Tag, multimodal bool, locale string) agent.Tool {
	return agent.Tool{
		Def: agent.ToolDef{
			Name:        "search_echos",
			Description: "检索用户过往发布的 Echo（微博客/碎碎念）。可用 query 做语义/关键词检索，并可选地用 tags（标签名）与 date_from/date_to（日期范围）做精确筛选；三者可组合，但至少提供其一。query 传精炼核心词，不要整句。",
			Parameters:  json.RawMessage(`{"type":"object","properties":{"query":{"type":"string","description":"语义/关键词检索词，传与问题最相关的核心词（精炼，不要整句）；仅按标签或时间筛选时可省略"},"tags":{"type":"array","items":{"type":"string"},"description":"按标签名筛选（标签名而非ID），如 [\"读书\",\"旅行\"]；可用标签见系统提示"},"date_from":{"type":"string","description":"起始日期，格式 YYYY-MM-DD，含当天"},"date_to":{"type":"string","description":"结束日期，格式 YYYY-MM-DD，含当天"}}}`),
		},
		Execute: func(ctx context.Context, args json.RawMessage) (agent.ToolOutput, error) {
			var a searchArgs
			_ = json.Unmarshal(args, &a)
			a.Query = strings.TrimSpace(a.Query)
			tagIDs := resolveTagIDs(allTags, a.Tags)
			from := parseDay(a.DateFrom, false)
			to := parseDay(a.DateTo, true)

			structured := len(tagIDs) > 0 || from > 0 || to > 0
			if a.Query == "" && !structured {
				return agent.ToolOutput{}, errors.New("检索需要 query、tags 或日期范围至少其一")
			}

			var results []embeddingModel.SearchResult
			var total int64
			var execErr error
			switch {
			case structured:
				// 带结构化过滤：SQL 精确路径，query 降级为 content LIKE。
				results, total, execErr = s.queryEchos(ctx, a.Query, tagIDs, from, to)
			case s.embedding.Enabled(ctx):
				results, execErr = s.embedding.Search(ctx, a.Query, defaultTopK)
			default:
				results, total, execErr = s.queryEchos(ctx, a.Query, nil, 0, 0)
			}
			if execErr != nil {
				return agent.ToolOutput{}, execErr
			}
			// 命中后回查：Extension 文本（常开，仅几字 token）+ 配图（多模态开关）一次加载取齐。
			exts, images := s.enrichHits(ctx, results, multimodal)
			content := formatSearchResults(results, exts)
			// 命中数多于本次展示（top-k 截断）时如实告知模型，避免它把「采样」当「全部」；
			// 若要覆盖全部用于总结，应改调 summarize_echos。
			if total > int64(len(results)) {
				content = searchCoverageNoteFor(locale, int(total), len(results)) + "\n" + content
			}
			return agent.ToolOutput{
				Content: content,
				Meta:    results,
				Images:  images,
			}, nil
		},
	}
}

// queryEchos 走 echoService.QueryEchos（SQL 检索 + 可见性由 viewer 上下文裁决：
// /chat 为 admin，故 showPrivate=true，与向量索引可见性一致）。结果映射成与向量检索
// 同一形状，上层（formatSearchResults / SSE sources）无需区分两条路径。
// 返回的 total 是区间内命中总数（QueryEchos.Total），用于在 top-k 截断时如实回报覆盖度。
func (s *CopilotService) queryEchos(ctx context.Context, search string, tagIDs []string, from, to int64) ([]embeddingModel.SearchResult, int64, error) {
	page, err := s.echoService.QueryEchos(ctx, commonModel.EchoQueryDto{
		Page:     1,
		PageSize: defaultTopK,
		Search:   search,
		TagIDs:   tagIDs,
		DateFrom: from,
		DateTo:   to,
	})
	if err != nil {
		return nil, 0, err
	}
	results := make([]embeddingModel.SearchResult, 0, len(page.Items))
	for i := range page.Items {
		results = append(results, echoToSearchResult(page.Items[i]))
	}
	return results, page.Total, nil
}

// resolveTagIDs 把模型给的标签名（大小写不敏感）解析成标签 ID；匹配不上的名静默忽略。
func resolveTagIDs(allTags []echoModel.Tag, names []string) []string {
	if len(names) == 0 {
		return nil
	}
	byName := make(map[string]string, len(allTags))
	for _, t := range allTags {
		byName[strings.ToLower(t.Name)] = t.ID
	}
	ids := make([]string, 0, len(names))
	for _, n := range names {
		if id, ok := byName[strings.ToLower(strings.TrimSpace(n))]; ok {
			ids = append(ids, id)
		}
	}
	return ids
}

// parseDay 把 YYYY-MM-DD 解析成 Unix 秒（UTC）；endOfDay 为真时取当天 23:59:59，
// 用于把闭区间的右端覆盖到整天。解析失败或空串返回 0（视为未设置）。
func parseDay(s string, endOfDay bool) int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return 0
	}
	if endOfDay {
		t = t.Add(24*time.Hour - time.Second)
	}
	return t.Unix()
}

// searchHintOf 从工具入参拼一条人读的检索提示（供 SSE searching 事件展示），
// 组合 query / #标签 / 日期范围。
func searchHintOf(args json.RawMessage) string {
	var a searchArgs
	_ = json.Unmarshal(args, &a)
	parts := make([]string, 0, 3)
	if q := strings.TrimSpace(a.Query); q != "" {
		parts = append(parts, q)
	}
	for _, t := range a.Tags {
		if t = strings.TrimSpace(t); t != "" {
			parts = append(parts, "#"+t)
		}
	}
	if from, to := strings.TrimSpace(a.DateFrom), strings.TrimSpace(a.DateTo); from != "" || to != "" {
		parts = append(parts, from+"~"+to)
	}
	return strings.Join(parts, " ")
}

// formatSearchResults 把检索命中拼成回喂模型的精简文本（文本快照 + 扩展分享，控制 token）。
// exts 是 echoID → Extension 渲染文本（来自 enrichHits），命中时补在内容之后。
func formatSearchResults(results []embeddingModel.SearchResult, exts map[string]string) string {
	if len(results) == 0 {
		return "（没有检索到相关的 Echo）"
	}
	var b strings.Builder
	for i, r := range results {
		day := time.Unix(r.EchoCreated, 0).UTC().Format("2006-01-02")
		parts := []string{fmt.Sprintf("【%d】(%s)", i+1, day)}
		if c := strings.TrimSpace(r.Content); c != "" {
			parts = append(parts, c)
		}
		if ext := exts[r.EchoID]; ext != "" {
			parts = append(parts, ext)
		}
		b.WriteString(strings.Join(parts, " "))
		b.WriteString("\n")
	}
	return strings.TrimSpace(b.String())
}
