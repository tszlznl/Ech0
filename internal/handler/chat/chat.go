// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package handler 暴露 Chat 与 Embedding 配置的 HTTP 接口。
package handler

import (
	"github.com/gin-gonic/gin"
	res "github.com/lin-snow/ech0/internal/handler/response"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	chatService "github.com/lin-snow/ech0/internal/service/chat"
	embeddingService "github.com/lin-snow/ech0/internal/service/embedding"
)

type ChatHandler struct {
	chatService      chatService.Service
	embeddingService embeddingService.Service
}

func NewChatHandler(
	chatService chatService.Service,
	embeddingService embeddingService.Service,
) *ChatHandler {
	return &ChatHandler{
		chatService:      chatService,
		embeddingService: embeddingService,
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

// GetEmbeddingSettings 获取 Embedding 配置。
func (chatHandler *ChatHandler) GetEmbeddingSettings() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		setting, err := chatHandler.embeddingService.GetSetting(ctx.Request.Context())
		if err != nil {
			return res.Response{Err: err}
		}
		return res.Response{Data: setting}
	})
}

// UpdateEmbeddingSettings 更新 Embedding 配置。
func (chatHandler *ChatHandler) UpdateEmbeddingSettings() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var dto settingModel.EmbeddingSettingDto
		if err := ctx.ShouldBindJSON(&dto); err != nil {
			return res.Response{Err: err}
		}
		if err := chatHandler.embeddingService.UpdateSetting(ctx.Request.Context(), dto); err != nil {
			return res.Response{Err: err}
		}
		return res.Response{}
	})
}

// Reindex 触发对全部 Echo 的向量索引回填（管理员）。
func (chatHandler *ChatHandler) Reindex() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		result, err := chatHandler.embeddingService.Backfill(ctx.Request.Context())
		if err != nil {
			return res.Response{Err: err}
		}
		return res.Response{Data: result}
	})
}
