// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package handler 暴露 Chat 流式问答的 HTTP 接口。
package handler

import (
	"github.com/gin-gonic/gin"
	chatService "github.com/lin-snow/ech0/internal/service/chat"
)

type ChatHandler struct {
	chatService chatService.Service
}

func NewChatHandler(chatService chatService.Service) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
	}
}

type askRequest struct {
	Question string `json:"question"`
}

// Ask 处理 Chat 流式问答（SSE）。错误以 SSE 事件回传，故此处忽略返回值。
func (chatHandler *ChatHandler) Ask() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req askRequest
		_ = ctx.ShouldBindJSON(&req)
		_ = chatHandler.chatService.AskStream(ctx.Request.Context(), req.Question, ctx.Writer)
	}
}
