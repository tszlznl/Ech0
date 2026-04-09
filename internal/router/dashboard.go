package router

import (
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/middleware"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
)

func setupDashboardRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	// Auth
	appRouterGroup.AuthRouterGroup.GET(
		"/system/check-update",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.DashboardHandler.CheckUpdate(),
	)
	appRouterGroup.AuthRouterGroup.GET(
		"/system/logs",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.DashboardHandler.GetSystemLogs(),
	)
	appRouterGroup.AuthRouterGroup.GET(
		"/system/logs/stream",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.DashboardHandler.SSESubscribeSystemLogs(),
	)
	appRouterGroup.WSRouterGroup.GET("/system/logs", h.DashboardHandler.WSSubscribeSystemLogs())
}
