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

	// 公开可读接口：注册到「可匿名降级」组——无 token / 无效 token 按匿名继续，
	// 携带有效 token 时按用户身份（管理员可见私密内容）。是否匿名由所在路由组决定，
	// 不再依赖中间件内部的 path 名单。
	appRouterGroup.OptionalAuthRouterGroup.POST("/echo/query", h.EchoHandler.QueryEchos())

	// Deprecated: 以下分页/标签查询接口保留向后兼容，请优先使用 POST /echo/query
	//nolint:staticcheck // SA1019: 兼容旧客户端
	appRouterGroup.OptionalAuthRouterGroup.GET("/echo/page", h.EchoHandler.GetEchosByPage())
	//nolint:staticcheck // SA1019: 兼容旧客户端
	appRouterGroup.OptionalAuthRouterGroup.POST("/echo/page", h.EchoHandler.GetEchosByPage())
	//nolint:staticcheck // SA1019: 兼容旧客户端
	appRouterGroup.OptionalAuthRouterGroup.GET("/echo/tag/:tagid", h.EchoHandler.GetEchosByTagId())

	appRouterGroup.OptionalAuthRouterGroup.GET("/echo/today", h.EchoHandler.GetTodayEchos())
	appRouterGroup.OptionalAuthRouterGroup.GET("/echo/hot", h.EchoHandler.GetHotEchos())
	appRouterGroup.OptionalAuthRouterGroup.GET("/echo/random", h.EchoHandler.GetRandomEcho())
	appRouterGroup.OptionalAuthRouterGroup.GET("/echo/onthisday", h.EchoHandler.GetOnThisDayEchos())
	appRouterGroup.OptionalAuthRouterGroup.GET("/echo/:id", h.EchoHandler.GetEchoById())

	// 写操作：强制鉴权 + scope 校验。
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
