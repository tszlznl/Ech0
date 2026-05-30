// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package agent

import (
	"context"
	"encoding/json"

	model "github.com/lin-snow/ech0/internal/model/setting"
)

// Role 表示一条对话消息的发送者角色
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool" // 工具执行结果（function calling）
)

// Message 是 Agent 内部使用的、tool-aware 的对话消息抽象。
type Message struct {
	Role    Role
	Content string
	// ToolCalls 仅 RoleAssistant：本条 assistant 消息发起的工具调用（用于回灌上下文）。
	ToolCalls []ToolCall
	// ToolCallID 仅 RoleTool：本条结果对应的 ToolCall.ID。
	ToolCallID string
}

// ToolCall 是模型发起的一次工具调用（各家 SDK 的分片在 Provider 内拼装完整后才上浮）。
type ToolCall struct {
	ID   string          // 各家 SDK 的调用 id
	Name string          // 工具名
	Args json.RawMessage // 入参（完整 JSON）
}

// ToolDef 是工具对模型暴露的声明（名称 + 描述 + JSON Schema）。
type ToolDef struct {
	Name        string
	Description string
	Parameters  json.RawMessage // JSON Schema
}

// Tool 把工具声明与执行闭包绑定；执行体由领域层（Copilot Service）注入，agent 包零领域依赖。
type Tool struct {
	Def     ToolDef
	Execute func(ctx context.Context, args json.RawMessage) (ToolOutput, error)
}

// ToolOutput 是工具执行结果：Content 回喂模型，Meta 旁路带出领域数据（如命中的检索结果，供 SSE sources）。
type ToolOutput struct {
	Content string
	Meta    any
}

// Request 是一次 Provider 调用的协议无关载荷。
type Request struct {
	Messages    []Message
	Tools       []ToolDef
	Temperature *float32
	MaxTokens   int
}

// Response 是一次非流式生成的结果（Generate 用，无工具）。
type Response struct {
	Text string
}

// EventKind 区分 Provider 上浮的语义事件类型。
type EventKind int

const (
	EventTextDelta EventKind = iota // 文本增量
	EventToolCall                   // 一个拼装完整的工具调用
	EventDone                       // 本次 Provider 调用结束
	EventError                      // 传输/协议错误（之后 channel 关闭）
)

// Event 是 Provider→Loop 的统一事件。Provider 不懂业务语义，只吐文本增量与工具调用。
type Event struct {
	Kind     EventKind
	Text     string   // EventTextDelta
	ToolCall ToolCall // EventToolCall
	Err      error    // EventError
}

// RunRequest 是 Loop 层对领域层（Copilot Service）暴露的请求。
type RunRequest struct {
	Setting   model.AgentSetting
	Messages  []Message
	Tools     []Tool
	MaxRounds int      // 0 → 默认 defaultMaxRounds
	Temp      *float32 // nil → 不设置
}

// AgentEventKind 区分 Loop 上浮给领域层的语义事件类型。
type AgentEventKind int

const (
	AgentDelta      AgentEventKind = iota // 文本上屏（跨轮连续）
	AgentSearching                        // 模型决定调用工具（含 name + args）
	AgentToolResult                       // 工具执行完（Meta 即 ToolOutput.Meta，供 sources）
	AgentDone                             // 收尾
	AgentError                            // 中止
)

// AgentEvent 是 Loop→Copilot Service 的统一事件；语义翻译（searching/sources）在此完成。
type AgentEvent struct {
	Kind     AgentEventKind
	Text     string          // AgentDelta
	ToolName string          // AgentSearching
	ToolArgs json.RawMessage // AgentSearching
	Meta     any             // AgentToolResult
	Err      error           // AgentError
}
