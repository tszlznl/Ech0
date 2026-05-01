// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import "github.com/lin-snow/ech0/internal/handler"

func setupInitRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	appRouterGroup.PublicRouterGroup.GET("/init/status", h.InitHandler.GetInitStatus())
	appRouterGroup.PublicRouterGroup.POST("/init/owner", h.InitHandler.InitOwner())
}
