// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/handler/humares"
	"github.com/lin-snow/ech0/internal/middleware"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	authService "github.com/lin-snow/ech0/internal/service/auth"
)

// registerEmbeddingHuma 注册 Embedding 向量索引操作（owner / 管理员，需 admin:settings scope）。
// 注意：Embedding 设置（get/update）仍归口到 setupSettingRoutes（尚未迁移）。
//
// 鉴权复用现有 gin 中间件，经 humares.Bridge 适配为 Huma operation 中间件——零分叉：
// 拒绝时由中间件直接写出与旧版完全一致的本地化 401/403 响应。
func registerEmbeddingHuma(api huma.API, h *handler.Bundle, revoker authService.TokenRevoker) {
	secured := huma.Middlewares{
		humares.Bridge(middleware.RequireAuth(revoker)),
		humares.Bridge(middleware.RequireScopes(authModel.ScopeAdminSettings)),
	}
	security := humares.Secured(authModel.ScopeAdminSettings)

	huma.Register(api, huma.Operation{
		OperationID: "embedding-reindex",
		Method:      http.MethodPost,
		Path:        "/embedding/reindex",
		Summary:     "触发全量向量索引重建",
		Description: "提交一次全量向量索引回填作业，起即返回（异步）。",
		Tags:        []string{"Embedding"},
		Security:    security,
		Middlewares: secured,
	}, h.EmbeddingHandler.Reindex)

	huma.Register(api, huma.Operation{
		OperationID: "embedding-reindex-status",
		Method:      http.MethodGet,
		Path:        "/embedding/reindex/status",
		Summary:     "查询重建索引作业状态",
		Description: "前端按类型轮询；查无作业行时返回 status=idle。",
		Tags:        []string{"Embedding"},
		Security:    security,
		Middlewares: secured,
	}, h.EmbeddingHandler.ReindexStatus)

	huma.Register(api, huma.Operation{
		OperationID: "embedding-reindex-cancel",
		Method:      http.MethodPost,
		Path:        "/embedding/reindex/cancel",
		Summary:     "取消进行中的重建索引作业",
		Description: "取消后返回最新状态（轮询收敛到 cancelled）。",
		Tags:        []string{"Embedding"},
		Security:    security,
		Middlewares: secured,
	}, h.EmbeddingHandler.CancelReindex)
}
