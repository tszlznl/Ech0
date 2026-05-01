// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package mcp

import (
	"github.com/gin-gonic/gin"
	agentService "github.com/lin-snow/ech0/internal/service/agent"
	commentService "github.com/lin-snow/ech0/internal/service/comment"
	commonService "github.com/lin-snow/ech0/internal/service/common"
	connectService "github.com/lin-snow/ech0/internal/service/connect"
	echoService "github.com/lin-snow/ech0/internal/service/echo"
	fileService "github.com/lin-snow/ech0/internal/service/file"
	settingService "github.com/lin-snow/ech0/internal/service/setting"
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
	agentSvc agentService.Service,
	settingSvc settingService.Service,
) *Handler {
	registry := NewRegistry()
	adapter := NewAdapter(echoSvc, userSvc, commentSvc, fileSvc, commonSvc, connectSvc, agentSvc, settingSvc)
	adapter.RegisterAll(registry)
	return &Handler{server: NewServer(registry)}
}

func (h *Handler) ServeEndpoint() gin.HandlerFunc {
	return func(c *gin.Context) {
		h.server.ServeHTTP(c.Writer, c.Request)
	}
}
