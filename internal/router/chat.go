// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/middleware"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
)

// setupChatRoutes 注册 Chat 与 Embedding 配置路由（全部 owner / 管理员）。
func setupChatRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	appRouterGroup.AuthRouterGroup.POST(
		"/chat",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.ChatHandler.Ask(),
	)
	appRouterGroup.AuthRouterGroup.GET(
		"/chat/embedding/settings",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.ChatHandler.GetEmbeddingSettings(),
	)
	appRouterGroup.AuthRouterGroup.PUT(
		"/chat/embedding/settings",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.ChatHandler.UpdateEmbeddingSettings(),
	)
	appRouterGroup.AuthRouterGroup.POST(
		"/chat/reindex",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.ChatHandler.Reindex(),
	)
}
