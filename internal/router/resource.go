// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"github.com/lin-snow/ech0/internal/handler"
)

// setupResourceRoutes 设置资源路由。
// API 文档已迁移到 Huma 内置 docs：/api/docs（spec：/api/openapi.json|.yaml）。
func setupResourceRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	appRouterGroup.ResourceGroup.GET("/robots.txt", h.CommonHandler.GetRobotsTxt)
	appRouterGroup.ResourceGroup.GET("/sitemap.xml", h.CommonHandler.GetSitemap)
	appRouterGroup.ResourceGroup.GET("/rss", h.CommonHandler.GetRss)
	appRouterGroup.ResourceGroup.GET("/healthz", h.CommonHandler.Healthz())
}
