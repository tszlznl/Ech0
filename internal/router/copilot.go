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

// setupCopilotRoutes 仅保留 Chat 流式问答（SSE）走裸 gin。
func setupCopilotRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	appRouterGroup.AuthRouterGroup.POST(
		"/chat",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.CopilotHandler.Ask(),
	)
}

// registerCopilotHuma 注册 Copilot 的 JSON 端点。
func registerCopilotHuma(api huma.API, h *handler.Bundle, revoker authService.TokenRevoker) {
	huma.Register(api, huma.Operation{
		OperationID: "copilot-recent",
		Method:      http.MethodGet,
		Path:        "/agent/recent",
		Summary:     "获取作者近况的 AI 总结",
		Tags:        []string{"Copilot"},
	}, h.CopilotHandler.GetRecent)

	sec := humares.Secured(authModel.ScopeAdminSettings)
	mw := securedMW(revoker, authModel.ScopeAdminSettings)

	huma.Register(api, huma.Operation{
		OperationID: "copilot-session-get",
		Method:      http.MethodGet,
		Path:        "/chat/session",
		Summary:     "获取持久化 Chat 会话",
		Tags:        []string{"Copilot"},
		Security:    sec,
		Middlewares: mw,
	}, h.CopilotHandler.GetSession)

	huma.Register(api, huma.Operation{
		OperationID: "copilot-session-clear",
		Method:      http.MethodDelete,
		Path:        "/chat/session",
		Summary:     "清除持久化 Chat 会话",
		Tags:        []string{"Copilot"},
		Security:    sec,
		Middlewares: mw,
	}, h.CopilotHandler.ClearSession)
}
