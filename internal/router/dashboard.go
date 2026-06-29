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

// setupDashboardRoutes 仅保留实时日志订阅走裸 gin：SSE 流 + WebSocket。
func setupDashboardRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	appRouterGroup.AuthRouterGroup.GET(
		"/system/logs/stream",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.DashboardHandler.SSESubscribeSystemLogs(),
	)
	appRouterGroup.WSRouterGroup.GET("/system/logs", h.DashboardHandler.WSSubscribeSystemLogs())
}

// registerDashboardHuma 注册仪表盘的 JSON 端点（admin:settings）。
func registerDashboardHuma(api huma.API, h *handler.Bundle, revoker authService.TokenRevoker) {
	sec := humares.Secured(authModel.ScopeAdminSettings)
	mw := securedMW(revoker, authModel.ScopeAdminSettings)

	huma.Register(api, huma.Operation{
		OperationID: "dashboard-check-update",
		Method:      http.MethodGet,
		Path:        "/system/check-update",
		Summary:     "检查 Ech0 版本更新",
		Tags:        []string{"Dashboard"},
		Security:    sec,
		Middlewares: mw,
	}, h.DashboardHandler.CheckUpdate)

	huma.Register(api, huma.Operation{
		OperationID: "dashboard-system-logs",
		Method:      http.MethodGet,
		Path:        "/system/logs",
		Summary:     "获取系统历史日志",
		Tags:        []string{"Dashboard"},
		Security:    sec,
		Middlewares: mw,
	}, h.DashboardHandler.GetSystemLogs)

	huma.Register(api, huma.Operation{
		OperationID: "dashboard-visitor-stats",
		Method:      http.MethodGet,
		Path:        "/system/visitor-stats",
		Summary:     "获取近七天访客统计",
		Tags:        []string{"Dashboard"},
		Security:    sec,
		Middlewares: mw,
	}, h.DashboardHandler.GetVisitorStats)
}
