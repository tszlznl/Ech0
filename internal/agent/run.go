// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package agent

import (
	"context"
	"strings"
	"unicode/utf8"

	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// defaultMaxRounds 是工具轮数上限护栏：防模型反复调工具死循环烧 token。
const defaultMaxRounds = 3

// maxParallelTools 是单轮内并发执行工具调用的上限：模型一轮发多个工具调用时并发跑（多为 I/O
// 密集的检索），削减串行延迟，同时 clamp 住并发度避免突发打满下游。
const maxParallelTools = 4

// defaultRunStrings 是 RunStrings 各字段留空时的回退（保持历史中文行为，向后兼容）。
var defaultRunStrings = RunStrings{
	DedupNote:       "（已检索过，结果见上）",
	UnknownTool:     "未知工具：",
	ToolError:       "工具执行失败：",
	ImageNote:       toolImageNote,
	ContextTrimNote: "（早前检索结果已省略以控制长度）",
}

// withDefaults 用 defaultRunStrings 填充留空字段。
func (s RunStrings) withDefaults() RunStrings {
	if s.DedupNote == "" {
		s.DedupNote = defaultRunStrings.DedupNote
	}
	if s.UnknownTool == "" {
		s.UnknownTool = defaultRunStrings.UnknownTool
	}
	if s.ToolError == "" {
		s.ToolError = defaultRunStrings.ToolError
	}
	if s.ImageNote == "" {
		s.ImageNote = defaultRunStrings.ImageNote
	}
	if s.ContextTrimNote == "" {
		s.ContextTrimNote = defaultRunStrings.ContextTrimNote
	}
	return s
}

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
	go func() {
		// per-run 超时护栏：Timeout>0 时给整轮（含工具循环）套个上限，防 provider 静默挂死；
		// <=0 沿用传入 ctx，行为不变。cancel 在 runLoop 收口（关闭 out）后触发。
		runCtx := ctx
		if req.Timeout > 0 {
			var cancel context.CancelFunc
			runCtx, cancel = context.WithTimeout(ctx, req.Timeout)
			defer cancel()
		}
		runLoop(runCtx, provider, req, out)
	}()
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
	strs := req.Strings.withDefaults()

	for round := 0; round < maxRounds; round++ {
		// 轮内 token 预算回收：超限时把最旧的工具结果替换为占位，防多轮累积撑爆窗口。
		trimContext(messages, req.MaxContextTokens, strs.ContextTrimNote)
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
		if !execTools(ctx, out, o.calls, toolByName, seen, &messages, strs) {
			return // ctx 取消
		}
	}

	// 工具轮用尽仍在调工具：强制一轮「不给工具」让模型据已检索到的结果作答，保证有回答。
	trimContext(messages, req.MaxContextTokens, strs.ContextTrimNote)
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

// execTools 执行一轮的工具调用：去重、emit Searching/ToolResult、把结果追加进 messages。
// 工具执行错误不中止（回喂模型自愈）；仅 ctx 取消时返回 false。
//
// 三段式（保留消息顺序、去重确定性，同时并发掉 I/O 密集的执行）：
//
//	A. 顺序预处理——去重命中 / 未知工具就地定好其 tool 结果消息，其余记为待执行；
//	B. 有界并发执行待执行项（emit Searching + Execute），结果按 index 写入各自槽，无竞态；
//	C. 顺序收尾——按调用原序 emit ToolResult、定好 tool 结果消息与（可选）带图消息。
//
// 追加顺序：**先把全部 tool 结果消息按原序追加，再追加带图 user 消息**。这样一轮 assistant 的
// 多个 tool_use 的 tool_result 紧邻聚合，满足 Anthropic「tool_result 必须在紧随的同一条 user
// 消息里与 tool_use 一一对应」的约束（旧逐条「结果→图→结果→图」会把后续 result 推远导致配对失败）。
func execTools(
	ctx context.Context,
	out chan<- AgentEvent,
	calls []ToolCall,
	toolByName map[string]Tool,
	seen map[string]bool,
	messages *[]Message,
	strs RunStrings,
) bool {
	n := len(calls)
	toolMsgs := make([]Message, n)   // 每个调用对应的 tool 结果消息（含去重/未知/错误/正常）
	imageMsgs := make([]*Message, n) // 每个调用可选的带图 user 消息（多模态）
	outputs := make([]ToolOutput, n)
	execErrs := make([]error, n)

	// A. 顺序预处理：去重与未知工具就地定好结果消息；其余记为待执行（保留原序 index）。
	var runnable []int
	for i, tc := range calls {
		key := tc.Name + ":" + string(tc.Args)
		if seen[key] {
			toolMsgs[i] = Message{Role: RoleTool, ToolCallID: tc.ID, Content: strs.DedupNote}
			continue
		}
		seen[key] = true
		if _, ok := toolByName[tc.Name]; !ok {
			toolMsgs[i] = Message{Role: RoleTool, ToolCallID: tc.ID, Content: strs.UnknownTool + tc.Name}
			continue
		}
		runnable = append(runnable, i)
	}

	// B. 有界并发执行：每个 goroutine emit Searching + Execute，结果写入独立 index 槽。
	// emit 失败（ctx 取消）→ 返回 ctx.Err() 让整组取消。g.Wait 阻塞至所有 goroutine 结束，
	// 故 outputs/execErrs 的写入在 Wait 返回前全部完成，后续顺序读取无竞态。
	var g errgroup.Group
	g.SetLimit(maxParallelTools)
	for _, idx := range runnable {
		idx, tc, tool := idx, calls[idx], toolByName[calls[idx].Name]
		g.Go(func() error {
			if !emit(ctx, out, AgentEvent{Kind: AgentSearching, ToolName: tc.Name, ToolArgs: tc.Args}) {
				return ctx.Err()
			}
			outputs[idx], execErrs[idx] = tool.Execute(ctx, tc.Args)
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return false // ctx 取消
	}

	// C. 顺序收尾：按原序 emit ToolResult、定好结果/带图消息。
	for _, idx := range runnable {
		tc := calls[idx]
		if execErrs[idx] != nil {
			logUtil.GetLogger().Warn("agent tool execute failed",
				zap.String("module", "agent"),
				zap.String("tool", tc.Name),
				zap.Error(execErrs[idx]))
			toolMsgs[idx] = Message{Role: RoleTool, ToolCallID: tc.ID, Content: strs.ToolError + execErrs[idx].Error()}
			continue
		}
		if !emit(ctx, out, AgentEvent{Kind: AgentToolResult, ToolName: tc.Name, Meta: outputs[idx].Meta}) {
			return false
		}
		toolMsgs[idx] = Message{Role: RoleTool, ToolCallID: tc.ID, Content: outputs[idx].Content}
		// 多模态：工具带出了图片（如命中 Echo 的配图）→ 用带图 user 消息递给模型。
		// 走 user 消息而非塞进 tool_result，是因 OpenAI 的 tool 角色消息只能纯文本，
		// user 带图两家协议都支持，一套逻辑通用。
		if len(outputs[idx].Images) > 0 {
			imageMsgs[idx] = &Message{Role: RoleUser, Content: strs.ImageNote, Images: outputs[idx].Images}
		}
	}

	// 先追加全部 tool 结果（聚合相邻，满足 Anthropic 配对约束），再追加带图消息。
	*messages = append(*messages, toolMsgs...)
	for i := range imageMsgs {
		if imageMsgs[i] != nil {
			*messages = append(*messages, *imageMsgs[i])
		}
	}
	return true
}

// trimContext 在轮内消息上下文超 budget 时回收最旧的工具结果：把其 Content 替换为 note 占位
// （保留消息与 ToolCallID 配对，绝不删消息——否则 tool_use/tool_result 失配会被 API 400）。
// budget<=0 时不回收。逐条替换直到回到预算内或没有可回收的工具结果。
func trimContext(messages []Message, budget int, note string) {
	if budget <= 0 {
		return
	}
	for contextTokens(messages) > budget {
		idx := -1
		for i := range messages {
			if messages[i].Role == RoleTool && messages[i].Content != note {
				idx = i
				break
			}
		}
		if idx < 0 {
			return // 没有可回收的工具结果了
		}
		messages[idx].Content = note
	}
}

// contextTokens 估算消息上下文的 token 总量（仅按文本 rune 计，图片不计）。
func contextTokens(messages []Message) int {
	total := 0
	for i := range messages {
		total += utf8.RuneCountInString(messages[i].Content)
	}
	return total
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
