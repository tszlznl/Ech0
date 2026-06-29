// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package handler 暴露 Ech0 Copilot 的 HTTP 接口：AI 近期总结与 Chat 会话（JSON，Huma）
// 以及 Chat 流式问答（SSE，裸 gin）。
package handler

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/handler/humares"
	i18n "github.com/lin-snow/ech0/internal/i18n"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	copilotService "github.com/lin-snow/ech0/internal/service/copilot"
	timezoneUtil "github.com/lin-snow/ech0/internal/util/timezone"
)

type CopilotHandler struct {
	summaryService copilotService.SummaryService
	chatService    copilotService.ChatService
}

func NewCopilotHandler(
	summaryService copilotService.SummaryService,
	chatService copilotService.ChatService,
) *CopilotHandler {
	return &CopilotHandler{
		summaryService: summaryService,
		chatService:    chatService,
	}
}

type (
	GetRecentInput    struct{}
	GetSessionInput   struct{}
	ClearSessionInput struct{}
)

// GetRecent 返回作者近况的 AI 总结（公开）。
func (h *CopilotHandler) GetRecent(ctx context.Context, _ *GetRecentInput) (*humares.Envelope[string], error) {
	gen, err := h.summaryService.GetRecent(ctx)
	if err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK(ctx, gen, commonModel.AGENT_GET_RECENT_SUCCESS), nil
}

// GetSession 返回当前登录用户的持久化 Chat 会话（admin:settings）。
func (h *CopilotHandler) GetSession(ctx context.Context, _ *GetSessionInput) (*humares.Envelope[[]copilotService.ChatMessage], error) {
	session, err := h.chatService.GetSession(ctx)
	if err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK(ctx, session, commonModel.CHAT_SESSION_GET_SUCCESS), nil
}

// ClearSession 删除当前登录用户的持久化 Chat 会话（admin:settings，不可恢复）。
func (h *CopilotHandler) ClearSession(ctx context.Context, _ *ClearSessionInput) (*humares.Envelope[any], error) {
	if err := h.chatService.ClearSession(ctx); err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK[any](ctx, nil, commonModel.CHAT_SESSION_CLEAR_SUCCESS), nil
}

type askRequest struct {
	Question string `json:"question"`
}

// Ask 处理 Chat 流式问答（SSE，裸 gin）。错误以 SSE 事件回传，故此处忽略返回值。
func (h *CopilotHandler) Ask() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req askRequest
		_ = ctx.ShouldBindJSON(&req)
		locale := i18n.LocaleFromGin(ctx)
		// 按用户上报时区算「今天/去年/上个月」与区间日界（与 today/heatmap 一致）。
		timezone := timezoneUtil.NormalizeTimezone(ctx.GetHeader(timezoneUtil.DefaultTimezoneHeader))
		_ = h.chatService.AskStream(ctx.Request.Context(), req.Question, locale, timezone, ctx.Writer)
	}
}
