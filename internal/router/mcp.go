package router

import (
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/middleware"
)

func setupMCPRoutes(groups *AppRouterGroup, h *handler.Bundle) {
	g := groups.MCPRouterGroup
	g.Use(middleware.RateLimit(20, 40), middleware.OriginGuard(nil))
	g.POST("", h.MCPHandler.ServeEndpoint())
	g.GET("", h.MCPHandler.ServeEndpoint())
}
