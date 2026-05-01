// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/middleware"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
)

// setupCommonRoutes 设置普通路由
func setupCommonRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	// Public
	appRouterGroup.PublicRouterGroup.GET("/heatmap", h.CommonHandler.GetHeatMap())
	appRouterGroup.PublicRouterGroup.GET("/hello", h.CommonHandler.HelloEch0())

	// Auth
	appRouterGroup.AuthRouterGroup.GET(
		"/backup/export",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.BackupHandler.ExportBackup(),
	)
	appRouterGroup.AuthRouterGroup.GET("/website/title", h.CommonHandler.GetWebsiteTitle())
}
