// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/middleware"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
)

// setupEchoRoutes 设置Echo路由
func setupEchoRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	// Public
	// 点赞接口保持匿名可访问，但叠加 IP 维度的限速 + (IP, echoID) 维度的去重窗口，
	// 防止匿名调用方反复刷 fav_count、放大数据库与缓存压力。窗口内的重复请求按
	// 幂等处理，返回与正常成功路径形状一致的响应。
	appRouterGroup.PublicRouterGroup.PUT(
		"/echo/like/:id",
		middleware.RateLimitWithIdempotency(2, 5, time.Hour, "id", func(c *gin.Context) {
			c.JSON(http.StatusOK, commonModel.OK[any](nil, commonModel.LIKE_ECHO_SUCCESS))
		}),
		h.EchoHandler.LikeEcho(),
	)
	appRouterGroup.PublicRouterGroup.GET("/tags", h.EchoHandler.GetAllTags())

	// Auth
	// 读接口保留“可匿名降级”行为：无 token 或无效 token 时由 JWT 中间件降级为匿名用户继续访问。
	appRouterGroup.AuthRouterGroup.POST("/echo/query", h.EchoHandler.QueryEchos())

	// Deprecated: 以下分页/标签查询接口保留向后兼容，请优先使用 POST /echo/query
	//nolint:staticcheck // SA1019: 兼容旧客户端
	appRouterGroup.AuthRouterGroup.GET("/echo/page", h.EchoHandler.GetEchosByPage())
	//nolint:staticcheck // SA1019: 兼容旧客户端
	appRouterGroup.AuthRouterGroup.POST("/echo/page", h.EchoHandler.GetEchosByPage())
	//nolint:staticcheck // SA1019: 兼容旧客户端
	appRouterGroup.AuthRouterGroup.GET("/echo/tag/:tagid", h.EchoHandler.GetEchosByTagId())

	appRouterGroup.AuthRouterGroup.GET("/echo/today", h.EchoHandler.GetTodayEchos())
	appRouterGroup.AuthRouterGroup.GET("/echo/hot", h.EchoHandler.GetHotEchos())
	appRouterGroup.AuthRouterGroup.GET("/echo/:id", h.EchoHandler.GetEchoById())
	appRouterGroup.AuthRouterGroup.POST(
		"/echo",
		middleware.RequireScopes(authModel.ScopeEchoWrite),
		h.EchoHandler.PostEcho(),
	)
	appRouterGroup.AuthRouterGroup.PUT(
		"/echo",
		middleware.RequireScopes(authModel.ScopeEchoWrite),
		h.EchoHandler.UpdateEcho(),
	)
	appRouterGroup.AuthRouterGroup.DELETE(
		"/echo/:id",
		middleware.RequireScopes(authModel.ScopeEchoWrite),
		h.EchoHandler.DeleteEcho(),
	)
	appRouterGroup.AuthRouterGroup.POST(
		"/tag",
		middleware.RequireScopes(authModel.ScopeEchoWrite),
		h.EchoHandler.CreateTag(),
	)
	appRouterGroup.AuthRouterGroup.DELETE(
		"/tag/:id",
		middleware.RequireScopes(authModel.ScopeEchoWrite),
		h.EchoHandler.DeleteTag(),
	)
	// appRouterGroup.AuthRouterGroup.PUT("/tag", h.EchoHandler.UpdateTag())
}
