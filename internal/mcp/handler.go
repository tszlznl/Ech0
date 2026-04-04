package mcp

import (
	"github.com/gin-gonic/gin"
	echoService "github.com/lin-snow/ech0/internal/service/echo"
	userService "github.com/lin-snow/ech0/internal/service/user"
)

type Handler struct {
	server *Server
}

func NewHandler(echoSvc echoService.Service, userSvc userService.Service) *Handler {
	registry := NewRegistry()
	adapter := NewAdapter(echoSvc, userSvc)
	adapter.RegisterAll(registry)
	return &Handler{server: NewServer(registry)}
}

func (h *Handler) ServeEndpoint() gin.HandlerFunc {
	return func(c *gin.Context) {
		h.server.ServeHTTP(c.Writer, c.Request)
	}
}
