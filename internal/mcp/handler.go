package mcp

import (
	"github.com/gin-gonic/gin"
	commentService "github.com/lin-snow/ech0/internal/service/comment"
	commonService "github.com/lin-snow/ech0/internal/service/common"
	connectService "github.com/lin-snow/ech0/internal/service/connect"
	echoService "github.com/lin-snow/ech0/internal/service/echo"
	fileService "github.com/lin-snow/ech0/internal/service/file"
	userService "github.com/lin-snow/ech0/internal/service/user"
)

type Handler struct {
	server *Server
}

func NewHandler(
	echoSvc echoService.Service,
	userSvc userService.Service,
	commentSvc commentService.Service,
	fileSvc fileService.Service,
	commonSvc commonService.Service,
	connectSvc connectService.Service,
) *Handler {
	registry := NewRegistry()
	adapter := NewAdapter(echoSvc, userSvc, commentSvc, fileSvc, commonSvc, connectSvc)
	adapter.RegisterAll(registry)
	return &Handler{server: NewServer(registry)}
}

func (h *Handler) ServeEndpoint() gin.HandlerFunc {
	return func(c *gin.Context) {
		h.server.ServeHTTP(c.Writer, c.Request)
	}
}
