package router

import (
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/middleware"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
)

// setupConnectRoutes 设置连接路由
func setupConnectRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	// Public
	appRouterGroup.PublicRouterGroup.GET("/connect", h.ConnectHandler.GetConnect())
	appRouterGroup.PublicRouterGroup.GET("/connect/list", h.ConnectHandler.GetConnects())
	appRouterGroup.PublicRouterGroup.GET("/connects/info", h.ConnectHandler.GetConnectsInfo())

	// Auth
	appRouterGroup.AuthRouterGroup.GET(
		"/connects/health",
		middleware.RequireScopes(authModel.ScopeConnectRead),
		h.ConnectHandler.GetConnectsHealth(),
	)
	appRouterGroup.AuthRouterGroup.POST(
		"/connects",
		middleware.RequireScopes(authModel.ScopeConnectWrite),
		h.ConnectHandler.AddConnect(),
	)
	appRouterGroup.AuthRouterGroup.DELETE(
		"/connects/:id",
		middleware.RequireScopes(authModel.ScopeConnectWrite),
		h.ConnectHandler.DeleteConnect(),
	)
}
