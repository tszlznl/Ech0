// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package agent

import (
	"context"
	"strings"

	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"go.uber.org/zap"
)

// defaultMaxRounds 是工具轮数上限护栏：防模型反复调工具死循环烧 token。
const defaultMaxRounds = 3

// Run 以 ReAct 结构（reason → act → observe）驱动一轮对话：模型在一次问答内
// 自主决定是否调用工具、调几次。文本增量实时上浮（AgentDelta），工具调用触发
// AgentSearching/AgentToolResult，收尾 AgentDone，错误 AgentError。
//
// 工具执行错误不中止——包装成 tool 结果回喂模型让它自愈；传输/协议错误才中止。
// 护栏：maxRounds 工具轮上限、同 turn 内查询去重、ctx 取消即停。工具轮用尽后
// 强制一轮「不给工具」的收尾，保证模型据已检索结果作答（否则会出现「只检索不回答」）。
func Run(ctx context.Context, req RunRequest) (<-chan AgentEvent, error) {
	if err := validate(req.Setting); err != nil {
		return nil, err
	}
	provider, err := providerFor(req.Setting)
	if err != nil {
		return nil, err
	}

	out := make(chan AgentEvent)
	go runLoop(ctx, provider, req, out)
	return out, nil
}

func runLoop(ctx context.Context, provider Provider, req RunRequest, out chan<- AgentEvent) {
	defer close(out)

	maxRounds := req.MaxRounds
	if maxRounds <= 0 {
		maxRounds = defaultMaxRounds
	}

	toolDefs := make([]ToolDef, 0, len(req.Tools))
	toolByName := make(map[string]Tool, len(req.Tools))
	for _, t := range req.Tools {
		toolDefs = append(toolDefs, t.Def)
		toolByName[t.Def.Name] = t
	}

	messages := req.Messages
	seen := make(map[string]bool)

	for round := 0; round < maxRounds; round++ {
		o := streamRound(ctx, provider, out, messages, toolDefs, req.Temp)
		if o.aborted {
			return // ctx 取消
		}
		if o.err != nil {
			emit(ctx, out, AgentEvent{Kind: AgentError, Err: o.err})
			return
		}
		if len(o.calls) == 0 {
			// 模型本轮直接作答（无工具调用）→ 正常收尾
			emit(ctx, out, AgentEvent{Kind: AgentDone})
			return
		}

		// 回灌本轮 assistant 的 tool_calls（连同已产出的文本），供下一轮上下文
		messages = append(messages, Message{Role: RoleAssistant, Content: o.assistant, ToolCalls: o.calls})
		if !execTools(ctx, out, o.calls, toolByName, seen, &messages) {
			return // ctx 取消
		}
	}

	// 工具轮用尽仍在调工具：强制一轮「不给工具」让模型据已检索到的结果作答，保证有回答。
	o := streamRound(ctx, provider, out, messages, nil, req.Temp)
	if o.aborted {
		return
	}
	if o.err != nil {
		emit(ctx, out, AgentEvent{Kind: AgentError, Err: o.err})
		return
	}
	emit(ctx, out, AgentEvent{Kind: AgentDone})
}

// roundOutcome 是一轮 provider.Stream 调用的结果。
type roundOutcome struct {
	calls     []ToolCall
	assistant string // 本轮产出的文本（已实时 emit，留作回灌上下文）
	aborted   bool   // ctx 取消
	err       error  // 传输/协议错误
}

// streamRound 跑一次 provider.Stream：文本增量实时 emit AgentDelta，收集工具调用。
// toolDefs 为 nil 时模型无可用工具，被迫直接作答（用于强制收尾轮）。
func streamRound(
	ctx context.Context,
	provider Provider,
	out chan<- AgentEvent,
	messages []Message,
	toolDefs []ToolDef,
	temp *float32,
) roundOutcome {
	evCh, err := provider.Stream(ctx, Request{
		Messages:    messages,
		Tools:       toolDefs,
		Temperature: temp,
	})
	if err != nil {
		return roundOutcome{err: err}
	}

	var (
		o roundOutcome
		b strings.Builder
	)
	for ev := range evCh {
		switch ev.Kind {
		case EventTextDelta:
			b.WriteString(ev.Text)
			if !emit(ctx, out, AgentEvent{Kind: AgentDelta, Text: ev.Text}) {
				o.aborted = true
			}
		case EventToolCall:
			o.calls = append(o.calls, ev.ToolCall)
		case EventError:
			o.err = ev.Err
		case EventDone:
			// 本次 Provider 调用结束，无需额外处理
		}
		if o.aborted {
			o.assistant = b.String()
			return o
		}
	}
	o.assistant = b.String()
	return o
}

// execTools 顺序执行一轮的工具调用：去重、emit Searching/ToolResult、把结果追加进 messages。
// 工具执行错误不中止（回喂模型自愈）；仅 ctx 取消时返回 false。
func execTools(
	ctx context.Context,
	out chan<- AgentEvent,
	calls []ToolCall,
	toolByName map[string]Tool,
	seen map[string]bool,
	messages *[]Message,
) bool {
	for _, tc := range calls {
		key := tc.Name + ":" + string(tc.Args)
		if seen[key] {
			*messages = append(*messages, Message{Role: RoleTool, ToolCallID: tc.ID, Content: "（已检索过，结果见上）"})
			continue
		}
		seen[key] = true

		tool, ok := toolByName[tc.Name]
		if !ok {
			*messages = append(*messages, Message{Role: RoleTool, ToolCallID: tc.ID, Content: "未知工具：" + tc.Name})
			continue
		}

		if !emit(ctx, out, AgentEvent{Kind: AgentSearching, ToolName: tc.Name, ToolArgs: tc.Args}) {
			return false
		}

		output, execErr := tool.Execute(ctx, tc.Args)
		if execErr != nil {
			logUtil.GetLogger().Warn("agent tool execute failed",
				zap.String("module", "agent"),
				zap.String("tool", tc.Name),
				zap.Error(execErr))
			*messages = append(*messages, Message{Role: RoleTool, ToolCallID: tc.ID, Content: "工具执行失败：" + execErr.Error()})
			continue
		}

		if !emit(ctx, out, AgentEvent{Kind: AgentToolResult, ToolName: tc.Name, Meta: output.Meta}) {
			return false
		}
		*messages = append(*messages, Message{Role: RoleTool, ToolCallID: tc.ID, Content: output.Content})

		// 多模态：工具带出了图片（如命中 Echo 的配图）→ 紧跟一条带图 user 消息递给模型。
		// 走 user 消息而非塞进 tool_result，是因 OpenAI 的 tool 角色消息只能纯文本，
		// user 带图两家协议都支持，一套逻辑通用。
		if len(output.Images) > 0 {
			*messages = append(*messages, Message{Role: RoleUser, Content: toolImageNote, Images: output.Images})
		}
	}
	return true
}

// toolImageNote 是带图 user 消息的说明文本，告诉模型这些图来自上一步检索命中的 Echo。
const toolImageNote = "（以下是上一步检索命中的 Echo 的配图，供你结合图片内容作答）"

// emit 向 out 发送事件，ctx 取消时返回 false（调用方应据此停止）。
func emit(ctx context.Context, out chan<- AgentEvent, ev AgentEvent) bool {
	select {
	case out <- ev:
		return true
	case <-ctx.Done():
		return false
	}
}
