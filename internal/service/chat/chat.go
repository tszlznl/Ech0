// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package service 实现基于 RAG 的 Chat（检索用户过往 Echo + 流式生成）。
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
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	embeddingModel "github.com/lin-snow/ech0/internal/model/embedding"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	embeddingService "github.com/lin-snow/ech0/internal/service/embedding"
)

// 检索默认返回的命中条数
const defaultTopK = 6

// Service 是 Chat 的对外接口。
type Service interface {
	AskStream(ctx context.Context, question string, w http.ResponseWriter) error
}

// KeyValueRepository 用于读取 Agent 生成配置。
type KeyValueRepository interface {
	GetKeyValue(ctx context.Context, key string) (string, error)
}

type ChatService struct {
	embedding embeddingService.Service
	kv        KeyValueRepository
}

var _ Service = (*ChatService)(nil)

func NewChatService(embedding embeddingService.Service, kv KeyValueRepository) *ChatService {
	return &ChatService{embedding: embedding, kv: kv}
}

func (s *ChatService) agentSetting(ctx context.Context) (settingModel.AgentSetting, error) {
	var setting settingModel.AgentSetting
	raw, err := s.kv.GetKeyValue(ctx, commonModel.AgentSettingKey)
	if err != nil {
		return setting, errors.New(commonModel.AGENT_SETTING_NOT_FOUND)
	}
	if err := json.Unmarshal([]byte(raw), &setting); err != nil {
		return setting, err
	}
	return setting, nil
}

// AskStream 执行检索 + 流式生成，并把全过程以 SSE 写入 w。
// 设计上：尽早写出 SSE 头，之后所有错误都以 SSE "error" 事件回传，而非 HTTP 状态码。
func (s *ChatService) AskStream(ctx context.Context, question string, w http.ResponseWriter) error {
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

	agentSetting, err := s.agentSetting(ctx)
	if err != nil {
		writeSSE(w, flusher, "error", map[string]string{"message": err.Error()})
		return nil
	}

	results, err := s.embedding.Search(ctx, question, defaultTopK)
	if err != nil {
		writeSSE(w, flusher, "error", map[string]string{"message": err.Error()})
		return nil
	}

	// 先回传引用来源，便于前端展示与跳转
	writeSSE(w, flusher, "sources", results)

	stream, err := agent.GenerateStream(ctx, agentSetting, buildMessages(question, results), false, 0.4)
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
		case chunk, ok := <-stream:
			if !ok {
				writeSSE(w, flusher, "done", map[string]bool{"done": true})
				return nil
			}
			if chunk.Err != nil {
				writeSSE(w, flusher, "error", map[string]string{"message": chunk.Err.Error()})
				return nil
			}
			if chunk.Delta != "" {
				writeSSE(w, flusher, "delta", map[string]string{"text": chunk.Delta})
			}
		}
	}
}

func writeSSE(w http.ResponseWriter, flusher http.Flusher, event string, data any) {
	payload, _ := json.Marshal(data)
	_, _ = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, payload)
	flusher.Flush()
}

func buildMessages(question string, results []embeddingModel.SearchResult) []agent.Message {
	var ctxBuilder strings.Builder
	for i, r := range results {
		day := time.Unix(r.EchoCreated, 0).UTC().Format("2006-01-02")
		fmt.Fprintf(&ctxBuilder, "【%d】(%s) %s\n", i+1, day, r.Content)
	}
	contextText := strings.TrimSpace(ctxBuilder.String())
	if contextText == "" {
		contextText = "（没有检索到相关的 Echo）"
	}

	system := `你是用户自己的 Echo（微博客/碎碎念）回顾助手。下面会给你检索到的若干条用户过往 Echo 作为上下文。
要求：
- 只依据提供的 Echo 内容作答，做跨条目的归纳、总结与回顾；
- 如果上下文里没有足够依据，就如实说明“你的 Echo 里没有相关记录”，不要编造；
- 用简洁自然的中文，可用 Emoji 和换行，不要输出 HTML 标签。`

	return []agent.Message{
		{Role: agent.RoleSystem, Content: system},
		{
			Role:    agent.RoleUser,
			Content: fmt.Sprintf("我的相关 Echo：\n%s\n\n我的问题：%s", contextText, question),
		},
	}
}
