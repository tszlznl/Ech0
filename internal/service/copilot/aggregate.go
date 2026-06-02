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
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	"github.com/lin-snow/ech0/internal/storage"
)

const (
	// aggregatePageSize 是区间取全时每页拉取的条数。取 100 对齐 echoService.QueryEchos 的页上限，
	// 一次拉满、减少翻页往返（collectRange 仍会按 total 续翻直到取尽）。
	aggregatePageSize = 100
	// maxAggregateEchos 是单次聚合纳入的硬上限：兜底防超大账号把一整年几万条全拉进内存/上下文。
	// 触顶按时间倒序保留「最近 N 条」，并在结果中如实标注截断，绝不静默丢弃。
	maxAggregateEchos = 5000
)

// summarizeArgs 是 summarize_echos 的入参：date_from/date_to 必填（区间），tags/focus 可选。
type summarizeArgs struct {
	DateFrom string   `json:"date_from"`
	DateTo   string   `json:"date_to"`
	Tags     []string `json:"tags"`
	Focus    string   `json:"focus"`
}

// aggregateCoverage 是 summarize_echos 的旁路覆盖度元数据（经 ToolOutput.Meta → SSE coverage 事件），
// 让模型与用户都清楚「这次总结覆盖了多少」，杜绝静默截断式的「看着很完整」。
type aggregateCoverage struct {
	Total     int64 `json:"total"`     // 区间内命中的 Echo 总数
	Returned  int   `json:"returned"`  // 实际纳入聚合的条数（受 maxAggregateEchos 约束）
	Buckets   int   `json:"buckets"`   // map-reduce 分块数；1 表示窗口放得下、直接整段塞入
	Truncated bool  `json:"truncated"` // 是否因硬上限截断（保留最近）
}

// summarizeEchosTool 是注入给 agent 的领域工具：穷举聚合某时间区间内的【全部】 Echo，
// 供模型撰写跨度较长的总结/回顾（年终/年度/季度/月度）。与 search_echos 的 top-k 互补——
// 它覆盖区间全部而非采样，并随窗口预算自适应（放得下整段塞入，放不下按体量分块 map-reduce）。
func (s *CopilotService) summarizeEchosTool(
	allTags []echoModel.Tag,
	setting settingModel.AgentSetting,
	locale string,
	loc *time.Location,
	user chatUser,
) agent.Tool {
	return agent.Tool{
		Def: agent.ToolDef{
			Name:        "summarize_echos",
			Description: "聚合某时间区间内的【全部】Echo，用于生成跨度较长的总结/回顾（如年终、年度、季度、月度总结）。它会覆盖区间内所有记录而非只采样几条，正是「帮我写年终/年度总结」这类需求该用的工具——这类请求请直接用它，不要先用 search_echos 采样。必须提供 date_from 与 date_to；可选 tags（按标签名限定主题）、focus（侧重点，如“工作”“读书”“心情”）。",
			Parameters:  json.RawMessage(`{"type":"object","properties":{"date_from":{"type":"string","description":"起始日期，格式 YYYY-MM-DD，含当天"},"date_to":{"type":"string","description":"结束日期，格式 YYYY-MM-DD，含当天"},"tags":{"type":"array","items":{"type":"string"},"description":"可选，按标签名限定主题；可用标签见系统提示"},"focus":{"type":"string","description":"可选，总结的侧重点，如“工作”“读书”“心情”"}},"required":["date_from","date_to"]}`),
		},
		Execute: func(ctx context.Context, args json.RawMessage) (agent.ToolOutput, error) {
			var a summarizeArgs
			_ = json.Unmarshal(args, &a)
			from := parseDay(a.DateFrom, false, loc)
			to := parseDay(a.DateTo, true, loc)
			if from == 0 && to == 0 {
				return agent.ToolOutput{}, errors.New("summarize_echos 需要 date_from 与 date_to 指定时间区间")
			}
			tagIDs := resolveTagIDs(allTags, a.Tags)

			echos, total, truncated, err := s.collectRange(ctx, user.ID, tagIDs, from, to)
			if err != nil {
				return agent.ToolOutput{}, err
			}
			reverseEchos(echos) // 倒序拉取后翻成时间正序，便于按月归纳与顺读成稿

			material, buckets, err := s.mapReduceSummary(ctx, setting, locale, echos, aggregateBudgetTokens(setting), loc)
			if err != nil {
				return agent.ToolOutput{}, err
			}

			cov := aggregateCoverage{
				Total:     total,
				Returned:  len(echos),
				Buckets:   buckets,
				Truncated: truncated,
			}
			header := aggregateMaterialHeaderFor(locale, int(total), len(echos), buckets, truncated)
			if focus := strings.TrimSpace(a.Focus); focus != "" {
				if localeIsZH(locale) {
					header += "（用户希望侧重：" + focus + "）"
				} else {
					header += " (User wants emphasis on: " + focus + ")"
				}
			}
			return agent.ToolOutput{
				Content: header + "\n\n" + material,
				Meta:    cov,
			}, nil
		},
	}
}

// collectRange 穷举拉取 [from,to]（可叠加 tagIDs）区间内的全部 Echo，与 search_echos 的
// top-k 不同——它要的是「覆盖全部」而非「最相关几条」。返回**完整 Echo 对象**（QueryEchos 已
// preload Extension/EchoFiles/Tags），让材料富化零额外查询。按 created_at 倒序（QueryEchos 默认）
// 分页累加，触顶 maxAggregateEchos 即停并标记 truncated（因倒序，保留的是最近的）。
func (s *CopilotService) collectRange(
	ctx context.Context,
	userID string,
	tagIDs []string,
	from, to int64,
) (echos []echoModel.Echo, total int64, truncated bool, err error) {
	for page := 1; ; page++ {
		res, qErr := s.echoService.QueryEchos(ctx, commonModel.EchoQueryDto{
			Page:     page,
			PageSize: aggregatePageSize,
			TagIDs:   tagIDs,
			DateFrom: from,
			DateTo:   to,
			UserID:   userID,
		})
		if qErr != nil {
			return nil, 0, false, qErr
		}
		total = res.Total
		for i := range res.Items {
			echos = append(echos, res.Items[i])
			if len(echos) >= maxAggregateEchos {
				return echos, total, total > int64(len(echos)), nil
			}
		}
		// 终止：空页（防御性，避免死循环）或已覆盖总数。不能用「本页未满」判断——
		// 服务层可能下调实际页大小，未满不代表取尽。
		if len(res.Items) == 0 || int64(len(echos)) >= total {
			return echos, total, false, nil
		}
	}
}

// reverseEchos 把倒序（最新在前）翻成时间正序（最早在前），便于按月分桶与顺读成稿。
func reverseEchos(echos []echoModel.Echo) {
	for l, r := 0, len(echos)-1; l < r; l, r = l+1, r-1 {
		echos[l], echos[r] = echos[r], echos[l]
	}
}

// mapReduceSummary 据窗口预算把区间内的 Echo 浓缩成「供模型写最终成稿」的中间材料：
//   - 放得下（estimateTokens ≤ budget）→ 直接返回**按月分块的完整材料**（buckets=1，大窗口在此吃满）；
//   - 放不下 → 按 **token 体量贪心分块**（而非按月，自然处理「某月很多某月很少」），每块 LLM map
//     成事实性摘要，拼接；若拼接仍超预算，再 reduce 一轮。
//
// map 阶段 v1 顺序执行（成本可预测、实现简单）；并行（bounded）留作后续优化。
// 任一 map/reduce 调用失败即上抛，由 agent loop 作为「工具执行失败」回喂模型自愈。
func (s *CopilotService) mapReduceSummary(
	ctx context.Context,
	setting settingModel.AgentSetting,
	locale string,
	echos []echoModel.Echo,
	budget int,
	loc *time.Location,
) (content string, buckets int, err error) {
	full := formatEchosByMonth(echos, loc)
	if estimateTokens(full) <= budget {
		return full, 1, nil
	}

	// 超预算：按体量贪心分块（重月份自动切多块、轻月份自动并块），每块一次 map。
	chunks := chunkEchosByBudget(echos, budget, loc)
	var b strings.Builder
	for _, ch := range chunks {
		digest, mErr := agent.Generate(ctx, setting, []agent.Message{
			{Role: agent.RoleSystem, Content: aggregateMapPromptFor(locale)},
			{Role: agent.RoleUser, Content: formatEchosByMonth(ch, loc)},
		}, false, nil)
		if mErr != nil {
			return "", 0, mErr
		}
		fmt.Fprintf(&b, "【%s】\n%s\n\n", dateSpanOf(ch, loc), strings.TrimSpace(digest))
	}
	joined := strings.TrimSpace(b.String())

	// 各块摘要拼接后仍超预算 → 再压一轮（保留每段要点与时间线）。
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

	return joined, len(chunks), nil
}

// chunkEchosByBudget 把时间正序的 Echo 按格式化后的 token 体量贪心切块，每块 ≤ budget。
// 这样「某月发得多、某月发得少」被天然抹平：密集时段切成多块、稀疏时段并入相邻块，
// 既不会让单块撑爆一次 map 调用，也不会为寥寥几条单独浪费一次调用。单条超 budget 时自成一块。
func chunkEchosByBudget(echos []echoModel.Echo, budget int, loc *time.Location) [][]echoModel.Echo {
	var chunks [][]echoModel.Echo
	var cur []echoModel.Echo
	curTokens := 0
	for _, e := range echos {
		t := estimateTokens(formatEchoLine(e, loc))
		if len(cur) > 0 && curTokens+t > budget {
			chunks = append(chunks, cur)
			cur = nil
			curTokens = 0
		}
		cur = append(cur, e)
		curTokens += t
	}
	if len(cur) > 0 {
		chunks = append(chunks, cur)
	}
	return chunks
}

// formatEchosByMonth 把时间正序的 Echo 按月分组成带小标题与计数的结构化材料，
// 比扁平流水更利于模型做均衡覆盖（不漏发得少的月份）。纯代码，无 LLM。
func formatEchosByMonth(echos []echoModel.Echo, loc *time.Location) string {
	if len(echos) == 0 {
		return "（该区间内没有 Echo）"
	}
	counts := make(map[string]int, 12)
	for i := range echos {
		counts[monthOf(echos[i], loc)]++
	}
	var b strings.Builder
	curMonth := ""
	for i := range echos {
		m := monthOf(echos[i], loc)
		if m != curMonth {
			if curMonth != "" {
				b.WriteString("\n")
			}
			fmt.Fprintf(&b, "## %s (%d)\n", m, counts[m])
			curMonth = m
		}
		b.WriteString(formatEchoLine(echos[i], loc))
		b.WriteString("\n")
	}
	return strings.TrimSpace(b.String())
}

// formatEchoLine 把单条 Echo 渲染成一行富化材料：日期 + 正文 + 标签 + 扩展分享 + 配图计数。
// 标签/扩展/配图都来自 QueryEchos 的 preload，零额外查询；纯图（无正文）条目也据此计入活跃度，
// 不再像裸文本那样贡献为零。
func formatEchoLine(e echoModel.Echo, loc *time.Location) string {
	day := time.Unix(e.CreatedAt, 0).In(loc).Format("2006-01-02")
	parts := []string{"(" + day + ")"}
	content := strings.TrimSpace(e.Content)
	if content != "" {
		parts = append(parts, content)
	}
	if tags := tagLabels(e.Tags); tags != "" {
		parts = append(parts, tags)
	}
	if ext := formatExtension(e.Extension); ext != "" {
		parts = append(parts, ext)
	}
	if n := imageCountOf(e); n > 0 {
		if content == "" {
			parts = append(parts, fmt.Sprintf("[img-only×%d]", n)) // 纯图无正文
		} else {
			parts = append(parts, fmt.Sprintf("[img×%d]", n))
		}
	}
	return strings.Join(parts, " ")
}

// tagLabels 把 Echo 的标签渲染成 "#标签1 #标签2"（高信号、几字 token），空标签忽略。
func tagLabels(tags []echoModel.Tag) string {
	if len(tags) == 0 {
		return ""
	}
	parts := make([]string, 0, len(tags))
	for _, t := range tags {
		if n := strings.TrimSpace(t.Name); n != "" {
			parts = append(parts, "#"+n)
		}
	}
	return strings.Join(parts, " ")
}

// imageCountOf 统计一条 Echo 的图片附件数（据 File.Category 判定，复用 storage 分类）。
func imageCountOf(e echoModel.Echo) int {
	n := 0
	for _, ef := range e.EchoFiles {
		if storage.NormalizeCategory(ef.File.Category).IsImageLike() {
			n++
		}
	}
	return n
}

// monthOf 取 Echo 创建时间的 YYYY-MM（按用户时区 loc）。
func monthOf(e echoModel.Echo, loc *time.Location) string {
	return time.Unix(e.CreatedAt, 0).In(loc).Format("2006-01")
}

// dateSpanOf 取一块 Echo 的日期跨度标签（首日 ~ 末日；同日则只显一日，按用户时区 loc）。
func dateSpanOf(echos []echoModel.Echo, loc *time.Location) string {
	if len(echos) == 0 {
		return ""
	}
	first := time.Unix(echos[0].CreatedAt, 0).In(loc).Format("2006-01-02")
	last := time.Unix(echos[len(echos)-1].CreatedAt, 0).In(loc).Format("2006-01-02")
	if first == last {
		return first
	}
	return first + " ~ " + last
}
