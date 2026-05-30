// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/middleware"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
)

// setupCopilotRoutes 注册 Ech0 Copilot 路由：
//   - GET  /agent/recent —— 作者近况 AI 总结（公开）。
//   - POST /chat        —— Chat 流式问答（owner / 管理员）。
//
// 路径保持与归并前一致，对外契约不变。
func setupCopilotRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	// Public
	appRouterGroup.PublicRouterGroup.GET("/agent/recent", h.CopilotHandler.GetRecent())

	// Auth (owner / admin)
	appRouterGroup.AuthRouterGroup.POST(
		"/chat",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.CopilotHandler.Ask(),
	)
	appRouterGroup.AuthRouterGroup.GET(
		"/chat/session",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.CopilotHandler.GetSession(),
	)
	appRouterGroup.AuthRouterGroup.DELETE(
		"/chat/session",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.CopilotHandler.ClearSession(),
	)
}
