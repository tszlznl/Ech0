// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package agent

import (
	"context"
	"encoding/json"
	"errors"
	"io"

	model "github.com/lin-snow/ech0/internal/model/setting"
	openai "github.com/sashabaranov/go-openai"
)

// openaiProvider 适配 OpenAI 兼容协议（OpenAI / DeepSeek / Qwen / Moonshot / Ollama 等）。
type openaiProvider struct {
	setting model.AgentSetting
}

func (p *openaiProvider) client() *openai.Client {
	cfg := openai.DefaultConfig(p.setting.ApiKey)
	if p.setting.BaseURL != "" {
		cfg.BaseURL = p.setting.BaseURL
	}
	return openai.NewClientWithConfig(cfg)
}

// buildMessages 把内部 tool-aware Message 映射为 OpenAI ChatCompletionMessage。
func (p *openaiProvider) buildMessages(in []Message) []openai.ChatCompletionMessage {
	msgs := make([]openai.ChatCompletionMessage, 0, len(in))
	for _, m := range in {
		msg := openai.ChatCompletionMessage{
			Role:       toOpenAIRole(m.Role),
			Content:    m.Content,
			ToolCallID: m.ToolCallID,
		}
		for _, tc := range m.ToolCalls {
			msg.ToolCalls = append(msg.ToolCalls, openai.ToolCall{
				ID:   tc.ID,
				Type: openai.ToolTypeFunction,
				Function: openai.FunctionCall{
					Name:      tc.Name,
					Arguments: string(tc.Args),
				},
			})
		}
		msgs = append(msgs, msg)
	}
	return msgs
}

// buildTools 把内部 ToolDef 映射为 OpenAI Tool（function calling）。
func (p *openaiProvider) buildTools(defs []ToolDef) []openai.Tool {
	if len(defs) == 0 {
		return nil
	}
	tools := make([]openai.Tool, 0, len(defs))
	for _, d := range defs {
		tools = append(tools, openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        d.Name,
				Description: d.Description,
				Parameters:  json.RawMessage(d.Parameters),
			},
		})
	}
	return tools
}

func (p *openaiProvider) Complete(ctx context.Context, req Request) (Response, error) {
	chatReq := openai.ChatCompletionRequest{
		Model:    p.setting.Model,
		Messages: p.buildMessages(req.Messages),
		Tools:    p.buildTools(req.Tools),
	}
	if req.Temperature != nil {
		chatReq.Temperature = *req.Temperature
	}
	if req.MaxTokens > 0 {
		chatReq.MaxTokens = req.MaxTokens
	}

	resp, err := p.client().CreateChatCompletion(ctx, chatReq)
	if err != nil {
		return Response{}, err
	}
	if len(resp.Choices) == 0 {
		return Response{}, errors.New("openai: empty response")
	}
	return Response{Text: resp.Choices[0].Message.Content}, nil
}

func (p *openaiProvider) Stream(ctx context.Context, req Request) (<-chan Event, error) {
	ch := make(chan Event)
	go p.stream(ctx, req, ch)
	return ch, nil
}

// stream 真流式消费 OpenAI SSE：文本增量实时 emit，工具调用按 index 跨 chunk 累积
// arguments 分片，拼装完整后 emit EventToolCall。
func (p *openaiProvider) stream(ctx context.Context, req Request, ch chan<- Event) {
	defer close(ch)

	chatReq := openai.ChatCompletionRequest{
		Model:    p.setting.Model,
		Messages: p.buildMessages(req.Messages),
		Tools:    p.buildTools(req.Tools),
		Stream:   true,
	}
	if req.Temperature != nil {
		chatReq.Temperature = *req.Temperature
	}
	if req.MaxTokens > 0 {
		chatReq.MaxTokens = req.MaxTokens
	}

	stream, err := p.client().CreateChatCompletionStream(ctx, chatReq)
	if err != nil {
		send(ctx, ch, Event{Kind: EventError, Err: err})
		return
	}
	defer func() { _ = stream.Close() }()

	// acc 按 tool_call 的 index 累积分片：id/name 首帧给出，arguments 跨帧拼接。
	acc := newToolCallAccumulator()

	for {
		resp, recvErr := stream.Recv()
		if errors.Is(recvErr, io.EOF) {
			break
		}
		if recvErr != nil {
			send(ctx, ch, Event{Kind: EventError, Err: recvErr})
			return
		}
		if len(resp.Choices) == 0 {
			continue
		}
		delta := resp.Choices[0].Delta

		if delta.Content != "" {
			if !send(ctx, ch, Event{Kind: EventTextDelta, Text: delta.Content}) {
				return
			}
		}
		acc.add(delta.ToolCalls)
	}

	// 流结束：把累积完整的工具调用依次上浮
	for _, tc := range acc.finish() {
		if !send(ctx, ch, Event{Kind: EventToolCall, ToolCall: tc}) {
			return
		}
	}
	send(ctx, ch, Event{Kind: EventDone})
}

// toolCallAccumulator 累积流式 tool_call 分片：OpenAI 把同一个调用的 arguments
// 按 index 跨多个 chunk 切片下发，需按 index 拼回完整 JSON。
type toolCallAccumulator struct {
	order []int
	byIdx map[int]*ToolCall
	args  map[int][]byte
}

func newToolCallAccumulator() *toolCallAccumulator {
	return &toolCallAccumulator{
		byIdx: make(map[int]*ToolCall),
		args:  make(map[int][]byte),
	}
}

func (a *toolCallAccumulator) add(deltas []openai.ToolCall) {
	for _, d := range deltas {
		idx := 0
		if d.Index != nil {
			idx = *d.Index
		}
		tc, ok := a.byIdx[idx]
		if !ok {
			tc = &ToolCall{}
			a.byIdx[idx] = tc
			a.order = append(a.order, idx)
		}
		if d.ID != "" {
			tc.ID = d.ID
		}
		if d.Function.Name != "" {
			tc.Name = d.Function.Name
		}
		if d.Function.Arguments != "" {
			a.args[idx] = append(a.args[idx], d.Function.Arguments...)
		}
	}
}

func (a *toolCallAccumulator) finish() []ToolCall {
	out := make([]ToolCall, 0, len(a.order))
	for _, idx := range a.order {
		tc := a.byIdx[idx]
		args := a.args[idx]
		if len(args) == 0 {
			args = []byte("{}")
		}
		out = append(out, ToolCall{ID: tc.ID, Name: tc.Name, Args: json.RawMessage(args)})
	}
	return out
}

func toOpenAIRole(r Role) string {
	switch r {
	case RoleSystem:
		return openai.ChatMessageRoleSystem
	case RoleAssistant:
		return openai.ChatMessageRoleAssistant
	case RoleTool:
		return openai.ChatMessageRoleTool
	default:
		return openai.ChatMessageRoleUser
	}
}

// send 向 ch 发送事件，ctx 取消时返回 false（Provider 据此停止）。
func send(ctx context.Context, ch chan<- Event, ev Event) bool {
	select {
	case ch <- ev:
		return true
	case <-ctx.Done():
		return false
	}
}
