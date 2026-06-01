// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/lin-snow/ech0/internal/agent"
	"github.com/lin-snow/ech0/internal/config"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	embeddingModel "github.com/lin-snow/ech0/internal/model/embedding"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	"github.com/lin-snow/ech0/pkg/viewer"
)

// chatTemperature 是 Chat 生成温度
const chatTemperature float32 = 0.4

func (s *CopilotService) agentSetting(ctx context.Context) (settingModel.AgentSetting, error) {
	var setting settingModel.AgentSetting
	raw, err := s.kvRepository.GetKeyValue(ctx, commonModel.AgentSettingKey)
	if err != nil {
		return setting, errors.New(commonModel.AGENT_SETTING_NOT_FOUND)
	}
	if err := json.Unmarshal([]byte(raw), &setting); err != nil {
		return setting, err
	}
	return setting, nil
}

// AskStream 以 Agent（function calling）形态执行一轮问答：模型在一次对话内自主决定
// 是否检索、检索几次（search_echos 工具），全过程以 SSE 写入 w。
//
// 设计上：尽早写出 SSE 头，之后所有错误都以 SSE "error" 事件回传，而非 HTTP 状态码。
// SSE 事件：searching（模型决定检索）/ sources（命中来源，可多次）/ delta（文本增量）/
// done（收尾）/ error（中止）。
func (s *CopilotService) AskStream(ctx context.Context, question string, locale string, w http.ResponseWriter) error {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return errors.New("streaming unsupported")
	}

	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("Connection", "keep-alive")
	h.Set("X-Accel-Buffering", "no")

	question = strings.TrimSpace(question)
	if question == "" {
		writeSSE(w, flusher, "error", map[string]string{"message": "empty question"})
		return nil
	}

	// 登录用户 ID：用于本轮问答正常收尾后把对话追加进该用户的持久化会话（仅展示恢复，模型不读历史）。
	userID := viewer.MustFromContext(ctx).UserID()
	// 收集本轮 assistant 文本与命中来源，正常收尾时一并持久化。
	var assistantBuf strings.Builder
	var collectedSources []embeddingModel.SearchResult

	agentSetting, err := s.agentSetting(ctx)
	if err != nil {
		writeSSE(w, flusher, "error", map[string]string{"message": err.Error()})
		return nil
	}

	// 一次性取标签：既注入 system prompt 供模型挑选，又供工具把标签名解析成 ID。
	allTags, _ := s.echoService.GetAllTags()
	today := time.Now().UTC().Format("2006-01-02")
	tagNames := tagNamesForPrompt(allTags)

	// 整请求 token 护栏：从历史预算里扣掉固定开销（system prompt + 工具定义），让历史让路，
	// 避免 system + 工具定义 + 多轮历史叠加撑爆上下文窗口（不足下限时至少保留最近若干轮）。
	historyBudget := max(maxHistoryTokens-estimateTokens(buildSystemPrompt(locale, today, tagNames))-toolDefTokenEstimate, minHistoryTokens)

	// 多轮记忆：加载已持久化的会话并投影成模型历史（在 persistTurn 之前加载，本轮 question
	// 不在其中，由 buildChatMessages 单独追加，不重复计入）。
	history := historyForModel(s.loadSession(ctx, userID), locale, historyBudget)

	temp := chatTemperature // 取地址需可寻址的局部变量（chatTemperature 是 const）
	stream, err := agent.Run(ctx, agent.RunRequest{
		Setting:  agentSetting,
		Messages: buildChatMessages(history, question, locale, today, tagNames),
		Tools: []agent.Tool{
			s.searchEchosTool(allTags, agentSetting.Multimodal, locale), // 点查：top-k 检索
			s.summarizeEchosTool(allTags, agentSetting, locale),         // 聚合：区间穷举 + 窗口自适应总结
		},
		Temp:    &temp,
		Strings: runStringsFor(locale),
		Timeout: time.Duration(config.Config().Agent.TimeoutSeconds) * time.Second,
	})
	if err != nil {
		writeSSE(w, flusher, "error", map[string]string{"message": err.Error()})
		return nil
	}

	keepAlive := time.NewTicker(15 * time.Second)
	defer keepAlive.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-keepAlive.C:
			_, _ = fmt.Fprint(w, ": keep-alive\n\n")
			flusher.Flush()
		case ev, ok := <-stream:
			if !ok {
				// Run 正常会在关闭前发 done/error；兜底以 done 收尾
				s.persistTurn(ctx, userID, question, assistantBuf.String(), collectedSources)
				writeSSE(w, flusher, "done", map[string]bool{"done": true})
				return nil
			}
			switch ev.Kind {
			case agent.AgentDelta:
				if ev.Text != "" {
					assistantBuf.WriteString(ev.Text)
					writeSSE(w, flusher, "delta", map[string]string{"text": ev.Text})
				}
			case agent.AgentSearching:
				writeSSE(w, flusher, "searching", map[string]string{
					"name":  ev.ToolName,
					"query": searchHintOf(ev.ToolArgs),
				})
			case agent.AgentToolResult:
				// 两类工具结果的 Meta 形状不同：search_echos → []SearchResult（sources 引用），
				// summarize_echos → aggregateCoverage（coverage 覆盖度）。按类型分流到各自 SSE 事件，
				// 既不把覆盖度当 sources 数组喂坏前端，也保持「加法不替换」（旧前端忽略未知 coverage）。
				switch meta := ev.Meta.(type) {
				case []embeddingModel.SearchResult:
					collectedSources = append(collectedSources, meta...)
					writeSSE(w, flusher, "sources", meta)
				case aggregateCoverage:
					writeSSE(w, flusher, "coverage", meta)
				}
			case agent.AgentDone:
				s.persistTurn(ctx, userID, question, assistantBuf.String(), collectedSources)
				writeSSE(w, flusher, "done", map[string]bool{"done": true})
				return nil
			case agent.AgentError:
				writeSSE(w, flusher, "error", map[string]string{"message": ev.Err.Error()})
				return nil
			}
		}
	}
}

func writeSSE(w http.ResponseWriter, flusher http.Flusher, event string, data any) {
	payload, _ := json.Marshal(data)
	_, _ = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, payload)
	flusher.Flush()
}
