// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/config"
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/middleware"
)

// RouterContext 聚合路由注册需要的上下文。
type RouterContext struct {
	Engine   *gin.Engine
	Handlers *handler.Bundle
	MWDeps   *middleware.Deps
	Groups   *AppRouterGroup
	// HumaAPI 是统一的 type-first OpenAPI 实例；各 register*Huma 把 JSON 端点注册其上。
	HumaAPI huma.API
}

// RouteModule 定义统一路由模块接口。
type RouteModule interface {
	Name() string
	Register(ctx *RouterContext)
}

type routeModule struct {
	name     string
	register func(ctx *RouterContext)
}

func (m routeModule) Name() string {
	return m.name
}

func (m routeModule) Register(ctx *RouterContext) {
	m.register(ctx)
}

func coreRouteModules() []RouteModule {
	return []RouteModule{
		routeModule{
			name: "template",
			register: func(ctx *RouterContext) {
				setupTemplateRoutes(ctx.Engine, ctx.Handlers)
			},
		},
		routeModule{
			name: "static-files",
			register: func(ctx *RouterContext) {
				cfg := config.Config().Storage
				root := cfg.DataRoot
				if root == "" {
					root = "data/files"
				}
				filesGroup := ctx.Engine.Group("api/files", middleware.StaticFileSecurity())
				filesGroup.StaticFS("/", http.Dir(root))
			},
		},
		routeModule{
			name: "middleware",
			register: func(ctx *RouterContext) {
				setupMiddleware(ctx.Engine)
			},
		},
		routeModule{
			name: "router-groups",
			register: func(ctx *RouterContext) {
				ctx.Groups = setupRouterGroup(ctx.Engine, ctx.MWDeps)
			},
		},
		// huma-api 必须在 "middleware" 之后创建，使 /api/docs、/api/openapi.* 继承全局
		// 中间件（Recovery / i18n / CORS 等）。auth/scope 由各 operation 自带的 Bridge 中间件处理。
		routeModule{
			name: "huma-api",
			register: func(ctx *RouterContext) {
				ctx.HumaAPI = setupHumaAPI(ctx.Engine)
			},
		},
	}
}

func featureRouteModules() []RouteModule {
	return []RouteModule{
		routeModule{
			name: "resource",
			register: func(ctx *RouterContext) {
				setupResourceRoutes(ctx.Groups, ctx.Handlers)
			},
		},
		routeModule{
			name: "auth",
			register: func(ctx *RouterContext) {
				setupAuthRoutes(ctx.Groups, ctx.Handlers)
			},
		},
		routeModule{
			name: "comment",
			register: func(ctx *RouterContext) {
				setupCommentRoutes(ctx.Groups, ctx.Handlers)
			},
		},
		routeModule{
			name: "file",
			register: func(ctx *RouterContext) {
				setupFileRoutes(ctx.Groups, ctx.Handlers)
			},
		},
		routeModule{
			name: "setting",
			register: func(ctx *RouterContext) {
				setupSettingRoutes(ctx.Groups, ctx.Handlers)
			},
		},
		routeModule{
			name: "dashboard",
			register: func(ctx *RouterContext) {
				setupDashboardRoutes(ctx.Groups, ctx.Handlers)
			},
		},
		routeModule{
			name: "copilot",
			register: func(ctx *RouterContext) {
				setupCopilotRoutes(ctx.Groups, ctx.Handlers)
			},
		},
		// huma-routes 注册所有已迁移到 type-first OpenAPI 的域（当前：embedding）。
		// 迁移新域时改 RegisterHumaOperations，无需在此新增模块。
		routeModule{
			name: "huma-routes",
			register: func(ctx *RouterContext) {
				RegisterHumaOperations(ctx.HumaAPI, ctx.Handlers, revokerFromCtx(ctx))
			},
		},
		routeModule{
			name: "migration",
			register: func(ctx *RouterContext) {
				setupMigrationRoutes(ctx.Groups, ctx.Handlers)
			},
		},
		routeModule{
			name: "mcp",
			register: func(ctx *RouterContext) {
				setupMCPRoutes(ctx.Groups, ctx.Handlers)
			},
		},
	}
}
