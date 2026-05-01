// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import "github.com/lin-snow/ech0/internal/handler"

func setupAgentRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	// Public
	appRouterGroup.PublicRouterGroup.GET("/agent/recent", h.AgentHandler.GetRecent())

	// Auth
}
