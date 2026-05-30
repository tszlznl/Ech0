// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package handler 暴露 Ech0 Copilot 的 HTTP 接口：AI 近期总结与 Chat 流式问答。
package handler

import (
	"github.com/gin-gonic/gin"
	res "github.com/lin-snow/ech0/internal/handler/response"
	i18n "github.com/lin-snow/ech0/internal/i18n"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	copilotService "github.com/lin-snow/ech0/internal/service/copilot"
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

// GetRecent 返回作者近况的 AI 总结。
func (h *CopilotHandler) GetRecent() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		gen, err := h.summaryService.GetRecent(ctx)
		if err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Data: gen,
			Msg:  commonModel.AGENT_GET_RECENT_SUCCESS,
		}
	})
}

// GetSession 返回当前登录用户的持久化 Chat 会话（重载页面恢复展示用）。
func (h *CopilotHandler) GetSession() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		session, err := h.chatService.GetSession(ctx.Request.Context())
		if err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Data: session,
			Msg:  commonModel.CHAT_SESSION_GET_SUCCESS,
		}
	})
}

// ClearSession 删除当前登录用户的持久化 Chat 会话（不可恢复，前端二次确认）。
func (h *CopilotHandler) ClearSession() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		if err := h.chatService.ClearSession(ctx.Request.Context()); err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Msg: commonModel.CHAT_SESSION_CLEAR_SUCCESS,
		}
	})
}

type askRequest struct {
	Question string `json:"question"`
}

// Ask 处理 Chat 流式问答（SSE）。错误以 SSE 事件回传，故此处忽略返回值。
func (h *CopilotHandler) Ask() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req askRequest
		_ = ctx.ShouldBindJSON(&req)
		locale := i18n.LocaleFromGin(ctx)
		_ = h.chatService.AskStream(ctx.Request.Context(), req.Question, locale, ctx.Writer)
	}
}
