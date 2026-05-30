// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package agent

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/packages/param"
	model "github.com/lin-snow/ech0/internal/model/setting"
)

// anthropicDefaultMaxTokens 是 Anthropic API 必填字段 max_tokens 的默认值
const anthropicDefaultMaxTokens = 4096

// anthropicProvider 适配 Anthropic（Claude）协议。
//
// Stream 是真流式：走 Messages.NewStreaming 逐事件消费 SSE，text_delta 实时
// 上浮为多个 EventTextDelta；工具调用的 input_json_delta 分片借 SDK 自带的
// Message.Accumulate 累积出完整 input，流结束后从累积好的 Message.Content 里逐个
// 上浮 EventToolCall。与 OpenAI provider 行为对齐，二者共用同一套 Run 循环与 SSE。
// Complete（非流式）仍走 generate，单次请求拿整段文本即可。
type anthropicProvider struct {
	setting model.AgentSetting
}

// newClient 构造带鉴权 / 自定义 BaseURL 的 Anthropic client（流式与非流式共用）。
func (p *anthropicProvider) newClient() anthropic.Client {
	opts := []option.RequestOption{option.WithAPIKey(p.setting.ApiKey)}
	if p.setting.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(p.setting.BaseURL))
	}
	return anthropic.NewClient(opts...)
}

// buildParams 把 Request 映射为 MessageNewParams。流式与非流式共用同一套
// model / max_tokens / system / tools / temperature 逻辑，避免两路实现漂移。
func (p *anthropicProvider) buildParams(req Request) anthropic.MessageNewParams {
	systemBlocks, msgs := p.buildMessages(req.Messages)

	maxTokens := int64(anthropicDefaultMaxTokens)
	if req.MaxTokens > 0 {
		maxTokens = int64(req.MaxTokens)
	}
	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(p.setting.Model),
		MaxTokens: maxTokens,
		Messages:  msgs,
	}
	if len(systemBlocks) > 0 {
		params.System = systemBlocks
	}
	if tools := p.buildTools(req.Tools); len(tools) > 0 {
		params.Tools = tools
	}
	if req.Temperature != nil {
		params.Temperature = param.NewOpt(float64(*req.Temperature))
	}
	return params
}

// generate 发起一次（非流式）Anthropic 请求，返回拼好的文本与解析出的工具调用。
func (p *anthropicProvider) generate(ctx context.Context, req Request) (string, []ToolCall, error) {
	client := p.newClient() // Messages.New 是指针方法，需绑定到可寻址的局部变量
	resp, err := client.Messages.New(ctx, p.buildParams(req))
	if err != nil {
		return "", nil, err
	}

	var text strings.Builder
	for _, block := range resp.Content {
		if block.Type == "text" {
			text.WriteString(block.Text)
		}
	}
	return text.String(), toolCallsFromContent(resp.Content), nil
}

// toolCallsFromContent 从 Message.Content（流式 Accumulate 后或一次性返回）里提取所有
// tool_use 调用。Args 为空时兜底成 "{}"，与 OpenAI accumulator.finish() 的约定一致。
func toolCallsFromContent(content []anthropic.ContentBlockUnion) []ToolCall {
	var calls []ToolCall
	for _, block := range content {
		if block.Type != "tool_use" {
			continue
		}
		args := json.RawMessage(block.Input)
		if len(args) == 0 {
			args = json.RawMessage("{}")
		}
		calls = append(calls, ToolCall{
			ID:   block.ID,
			Name: block.Name,
			Args: args,
		})
	}
	return calls
}

// buildMessages 把内部 tool-aware Message 映射为 Anthropic 的 system 块与消息序列。
// 连续的 RoleTool（同一 assistant 轮的多个工具结果）合并进单个 user 消息，
// 以满足 Anthropic「tool_result 必须与前一条 assistant 的 tool_use 一一对应且同处一条 user 消息」的约束。
func (p *anthropicProvider) buildMessages(in []Message) ([]anthropic.TextBlockParam, []anthropic.MessageParam) {
	var (
		systemBlocks []anthropic.TextBlockParam
		msgs         []anthropic.MessageParam
		pendingTools []anthropic.ContentBlockParamUnion
	)

	flush := func() {
		if len(pendingTools) > 0 {
			msgs = append(msgs, anthropic.NewUserMessage(pendingTools...))
			pendingTools = nil
		}
	}

	for _, m := range in {
		switch m.Role {
		case RoleSystem:
			flush()
			systemBlocks = append(systemBlocks, anthropic.TextBlockParam{Text: m.Content})
		case RoleAssistant:
			flush()
			var blocks []anthropic.ContentBlockParamUnion
			if m.Content != "" {
				blocks = append(blocks, anthropic.NewTextBlock(m.Content))
			}
			for _, tc := range m.ToolCalls {
				blocks = append(blocks, anthropic.NewToolUseBlock(tc.ID, json.RawMessage(tc.Args), tc.Name))
			}
			if len(blocks) > 0 {
				msgs = append(msgs, anthropic.NewAssistantMessage(blocks...))
			}
		case RoleTool:
			pendingTools = append(pendingTools, anthropic.NewToolResultBlock(m.ToolCallID, m.Content, false))
		default: // RoleUser
			flush()
			msgs = append(msgs, anthropic.NewUserMessage(anthropic.NewTextBlock(m.Content)))
		}
	}
	flush()
	return systemBlocks, msgs
}

// buildTools 把内部 ToolDef 映射为 Anthropic ToolUnionParam。
func (p *anthropicProvider) buildTools(defs []ToolDef) []anthropic.ToolUnionParam {
	if len(defs) == 0 {
		return nil
	}
	tools := make([]anthropic.ToolUnionParam, 0, len(defs))
	for _, d := range defs {
		var parsed struct {
			Properties any      `json:"properties"`
			Required   []string `json:"required"`
		}
		_ = json.Unmarshal(d.Parameters, &parsed)

		t := anthropic.ToolUnionParamOfTool(
			anthropic.ToolInputSchemaParam{
				Properties: parsed.Properties,
				Required:   parsed.Required,
			},
			d.Name,
		)
		if t.OfTool != nil && d.Description != "" {
			t.OfTool.Description = param.NewOpt(d.Description)
		}
		tools = append(tools, t)
	}
	return tools
}

func (p *anthropicProvider) Complete(ctx context.Context, req Request) (Response, error) {
	text, _, err := p.generate(ctx, req)
	if err != nil {
		return Response{}, err
	}
	if text == "" {
		return Response{}, errors.New("anthropic: empty text response")
	}
	return Response{Text: text}, nil
}

func (p *anthropicProvider) Stream(ctx context.Context, req Request) (<-chan Event, error) {
	ch := make(chan Event)
	go p.stream(ctx, req, ch)
	return ch, nil
}

// stream 真流式消费 Anthropic SSE：text_delta 增量实时 emit；tool_use 的
// input_json_delta 分片借 SDK 自带的 Message.Accumulate 累积出完整 input，
// 流结束后从累积好的 Message.Content 里逐个上浮 EventToolCall。
//
// 用 Accumulate 而非手动按 content block index 拼分片：SDK 内部已正确处理
// content_block_start / delta / stop 的状态机与 input 拼接，复用最稳，也免去一份
// 易错的手写累积逻辑。
func (p *anthropicProvider) stream(ctx context.Context, req Request, ch chan<- Event) {
	defer close(ch)

	client := p.newClient() // Messages.NewStreaming 是指针方法，需绑定到可寻址的局部变量
	stream := client.Messages.NewStreaming(ctx, p.buildParams(req))

	var acc anthropic.Message
	for stream.Next() {
		event := stream.Current()
		if err := acc.Accumulate(event); err != nil {
			send(ctx, ch, Event{Kind: EventError, Err: err})
			return
		}

		// 只有文本增量需要实时上浮；tool_use 分片交给 Accumulate，流结束后再统一吐出。
		if delta, ok := event.AsAny().(anthropic.ContentBlockDeltaEvent); ok {
			if td, ok := delta.Delta.AsAny().(anthropic.TextDelta); ok && td.Text != "" {
				if !send(ctx, ch, Event{Kind: EventTextDelta, Text: td.Text}) {
					return
				}
			}
		}
	}

	// 流结束：先看传输 / 协议错误（不静默吞错），再上浮累积好的工具调用，最后收尾。
	if err := stream.Err(); err != nil {
		send(ctx, ch, Event{Kind: EventError, Err: err})
		return
	}
	for _, tc := range toolCallsFromContent(acc.Content) {
		if !send(ctx, ch, Event{Kind: EventToolCall, ToolCall: tc}) {
			return
		}
	}
	send(ctx, ch, Event{Kind: EventDone})
}
