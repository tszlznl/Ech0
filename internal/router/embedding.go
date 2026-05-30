// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/middleware"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
)

// setupEmbeddingRoutes 注册 Embedding 向量索引操作路由（owner / 管理员）。
// 注意：Embedding 设置（get/update）归口到 setupSettingRoutes。
func setupEmbeddingRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	appRouterGroup.AuthRouterGroup.POST(
		"/embedding/reindex",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.EmbeddingHandler.Reindex(),
	)
}
