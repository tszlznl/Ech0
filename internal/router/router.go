// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/config"
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/middleware"
	authService "github.com/lin-snow/ech0/internal/service/auth"
)

type AppRouterGroup struct {
	ResourceGroup           *gin.RouterGroup
	PublicRouterGroup       *gin.RouterGroup
	AuthRouterGroup         *gin.RouterGroup
	OptionalAuthRouterGroup *gin.RouterGroup
	WSRouterGroup           *gin.RouterGroup
	MCPRouterGroup          *gin.RouterGroup
}

// SetupRouter 配置全部路由。装配显式分两段：
//
//  1. 核心（顺序敏感）：模板 → 静态文件 → 全局中间件 → 路由分组 → Huma API。
//     其中 Huma API 必须在全局中间件**之后**创建，使 /api/docs、/api/openapi.* 继承 Recovery/i18n/CORS。
//  2. 业务域路由：各域的裸 gin 端点（SSE/WS/上传/下载/captcha）+ RegisterHumaOperations 注册的 JSON 端点。
func SetupRouter(r *gin.Engine, h *handler.Bundle, mwDeps *middleware.Deps) {
	// 1. 核心
	setupTemplateRoutes(r, h)
	setupStaticFiles(r)
	setupMiddleware(r)
	groups := setupRouterGroup(r, mwDeps)
	api := setupHumaAPI(r)

	// 2. 业务域
	revoker := revokerOf(mwDeps)
	setupResourceRoutes(groups, h)
	setupAuthRoutes(groups, h)
	setupCommentRoutes(groups, h)
	setupFileRoutes(groups, h)
	setupDashboardRoutes(groups, h)
	setupCopilotRoutes(groups, h)
	RegisterHumaOperations(api, h, revoker) // 所有已迁移到 Huma 的 JSON 端点
	setupMigrationRoutes(groups, h)
	setupMCPRoutes(groups, h)
}

// setupStaticFiles 挂载本地上传文件的静态服务（/api/files），带目录穿越防护。
func setupStaticFiles(r *gin.Engine) {
	root := config.Config().Storage.DataRoot
	if root == "" {
		root = "data/files"
	}
	r.Group("api/files", middleware.StaticFileSecurity()).StaticFS("/", http.Dir(root))
}

// revokerOf 从中间件依赖里取出 token 吊销器（供鉴权 posture 复用 RequireAuth）。
func revokerOf(mwDeps *middleware.Deps) authService.TokenRevoker {
	if mwDeps != nil {
		return mwDeps.TokenRevoker
	}
	return nil
}

// setupRouterGroup 初始化路由组
func setupRouterGroup(r *gin.Engine, mwDeps *middleware.Deps) *AppRouterGroup {
	revoker := revokerOf(mwDeps)

	resource := r.Group("/")
	public := r.Group("/api")
	// 强制鉴权组：缺失/无效 token 一律 401。
	auth := r.Group("/api")
	auth.Use(middleware.NoCache(), middleware.RequireAuth(revoker))
	// 可匿名降级组：公开可读、但携带有效 token 时按用户身份（管理员见更多）。
	optionalAuth := r.Group("/api")
	optionalAuth.Use(middleware.NoCache(), middleware.OptionalAuth(revoker))
	ws := r.Group("/ws")
	mcpGroup := r.Group("/mcp")
	mcpGroup.Use(middleware.NoCache(), middleware.RequireAuth(revoker))
	return &AppRouterGroup{
		ResourceGroup:           resource,
		PublicRouterGroup:       public,
		AuthRouterGroup:         auth,
		OptionalAuthRouterGroup: optionalAuth,
		WSRouterGroup:           ws,
		MCPRouterGroup:          mcpGroup,
	}
}
