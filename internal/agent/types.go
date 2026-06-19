// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package agent

import (
	"context"
	"encoding/json"
	"time"

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
	// Images 仅 RoleUser：随本条消息发给多模态模型的图片（如检索命中 Echo 的配图）。
	// 普通文本消息为空；Provider 据此走多模态消息体（OpenAI multi-part / Anthropic image block）。
	Images []ImagePart
}

// ImagePart 是一张随消息发给多模态模型的图片。优先 Base64（自部署/私有存储下 provider 拉不到
// 内网 URL）；URL 仅用于可公开访问的 external 直链。二者取其一。
type ImagePart struct {
	MediaType string // MIME 类型，如 image/png、image/jpeg
	Base64    string // 图片字节的 base64（不含 data: 前缀）
	URL       string // 可公开访问的直链（external 存储用）
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
// Images 非空时，Loop 会在本条工具结果之后追加一条带图的 user 消息（多模态场景，如把命中 Echo 的配图递给模型）。
type ToolOutput struct {
	Content string
	Meta    any
	Images  []ImagePart
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
	EventTextDelta      EventKind = iota // 文本增量（答案正文）
	EventReasoningDelta                  // 推理增量（reasoning，与答案正文分流，仅供折叠展示）
	EventToolCall                        // 一个拼装完整的工具调用
	EventDone                            // 本次 Provider 调用结束
	EventError                           // 传输/协议错误（之后 channel 关闭）
)

// Event 是 Provider→Loop 的统一事件。Provider 不懂业务语义，只吐文本增量、推理增量与工具调用。
type Event struct {
	Kind     EventKind
	Text     string   // EventTextDelta / EventReasoningDelta
	ToolCall ToolCall // EventToolCall
	Err      error    // EventError
}

// RunStrings 是 Loop 在工具循环中回喂给模型 / 注入消息的少量提示文案。由领域层（知道 locale）
// 注入，使 agent 包保持 i18n 零依赖；任一字段留空则回退到 defaultRunStrings（中文，保持历史行为）。
type RunStrings struct {
	DedupNote       string // 同一查询重复调用时整条 tool 结果的内容
	UnknownTool     string // 未知工具名提示的前缀（后接工具名）
	ToolError       string // 工具执行失败提示的前缀（后接错误信息）
	ImageNote       string // 带图 user 消息的说明文本
	ContextTrimNote string // 轮内 token 预算回收时，替换最旧工具结果内容的占位文案
}

// RunRequest 是 Loop 层对领域层（Copilot Service）暴露的请求。
type RunRequest struct {
	Setting   model.AgentSetting
	Messages  []Message
	Tools     []Tool
	MaxRounds int           // 0 → 默认 defaultMaxRounds
	Temp      *float32      // nil → 不设置
	Strings   RunStrings    // 回喂/注入文案；零值字段回退到 defaultRunStrings
	Timeout   time.Duration // 单轮运行（含工具循环）整体超时；<=0 → 不额外设超时（沿用传入 ctx）
	// MaxContextTokens 是工具循环里整轮消息上下文的软上限（估算 token）；>0 时超限即回收
	// 最旧的工具结果（替换为 Strings.ContextTrimNote），防多轮工具结果累积撑爆窗口。0 → 不回收。
	MaxContextTokens int
}

// AgentEventKind 区分 Loop 上浮给领域层的语义事件类型。
type AgentEventKind int

const (
	AgentDelta      AgentEventKind = iota // 文本上屏（跨轮连续）
	AgentReasoning                        // 推理上屏（reasoning，与答案分流，不入答案/不回灌模型）
	AgentSearching                        // 模型决定调用工具（含 name + args）
	AgentToolResult                       // 工具执行完（Meta 即 ToolOutput.Meta，供 sources）
	AgentDone                             // 收尾
	AgentError                            // 中止
)

// AgentEvent 是 Loop→Copilot Service 的统一事件；语义翻译（searching/sources）在此完成。
type AgentEvent struct {
	Kind     AgentEventKind
	Text     string          // AgentDelta / AgentReasoning
	ToolName string          // AgentSearching
	ToolArgs json.RawMessage // AgentSearching
	Meta     any             // AgentToolResult
	Err      error           // AgentError
}
