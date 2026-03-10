package router

import "github.com/lin-snow/ech0/internal/handler"

func setupDashboardRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	// Auth
	appRouterGroup.AuthRouterGroup.GET("/dashboard/metrics", h.DashboardHandler.GetMetrics())
	appRouterGroup.AuthRouterGroup.GET("/system/logs", h.DashboardHandler.GetSystemLogs())
	appRouterGroup.AuthRouterGroup.GET("/system/logs/stream", h.DashboardHandler.SSESubscribeSystemLogs())
	appRouterGroup.WSRouterGroup.GET("/dashboard/metrics", h.DashboardHandler.WSSubsribeMetrics())
	appRouterGroup.WSRouterGroup.GET("/system/logs", h.DashboardHandler.WSSubscribeSystemLogs())
}
