// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package service 实现 Ech0 Copilot 域：AI 近期总结（summary）与基于 RAG 的
// Chat 流式问答（chat）。二者同属「Ech0 Copilot」产品域，共享 LLM 客户端
// （internal/agent）与检索能力，故归并在同一包内、按文件分隔关注点。
package service

import (
	"context"
	"net/http"

	echoService "github.com/lin-snow/ech0/internal/service/echo"
	embeddingService "github.com/lin-snow/ech0/internal/service/embedding"
	settingService "github.com/lin-snow/ech0/internal/service/setting"
)

// SummaryService 暴露 AI 近期总结能力（实现见 summary.go）。
type SummaryService interface {
	GetRecent(ctx context.Context) (string, error)
}

// ChatService 暴露 Chat 流式问答能力（实现见 chat.go）与会话持久化（实现见 session.go）。
type ChatService interface {
	AskStream(ctx context.Context, question string, locale string, w http.ResponseWriter) error
	// GetSession 返回当前登录用户的持久化会话（无会话时返回空切片）。
	GetSession(ctx context.Context) ([]ChatMessage, error)
	// ClearSession 删除当前登录用户的持久化会话。
	ClearSession(ctx context.Context) error
}

type (
	SettingService   = settingService.Service
	EchoService      = echoService.Service
	EmbeddingService = embeddingService.Service
)

// KeyValueRepository 用于读取 Agent 配置与读写近期总结缓存。
type KeyValueRepository interface {
	GetKeyValue(ctx context.Context, key string) (string, error)
	AddOrUpdateKeyValue(ctx context.Context, key, value string) error
	DeleteKeyValue(ctx context.Context, key string) error
}
