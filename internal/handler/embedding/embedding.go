// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package handler 暴露 Embedding 向量索引的 HTTP 接口。
//
// Embedding 的设置（get/update）归口到 setting 域；此处仅保留索引相关的操作类接口。
package handler

import (
	"github.com/gin-gonic/gin"
	res "github.com/lin-snow/ech0/internal/handler/response"
	embeddingService "github.com/lin-snow/ech0/internal/service/embedding"
)

type EmbeddingHandler struct {
	embeddingService embeddingService.Service
}

func NewEmbeddingHandler(embeddingService embeddingService.Service) *EmbeddingHandler {
	return &EmbeddingHandler{
		embeddingService: embeddingService,
	}
}

// Reindex 触发对全部 Echo 的向量索引回填（管理员）。
func (embeddingHandler *EmbeddingHandler) Reindex() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		result, err := embeddingHandler.embeddingService.Backfill(ctx.Request.Context())
		if err != nil {
			return res.Response{Err: err}
		}
		return res.Response{Data: result}
	})
}
