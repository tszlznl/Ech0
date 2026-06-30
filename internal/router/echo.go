// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/handler/humares"
	"github.com/lin-snow/ech0/internal/middleware"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	authService "github.com/lin-snow/ech0/internal/service/auth"
)

// registerEchoHuma 注册 Echo / Tag 路由（全部 JSON，已无裸 gin 端点）。
func registerEchoHuma(api huma.API, h *handler.Bundle, revoker authService.TokenRevoker) {
	optional := optionalMW(revoker)
	write := securedMW(revoker, authModel.ScopeEchoWrite)
	writeSec := humares.Secured(authModel.ScopeEchoWrite)

	// 点赞：匿名可访问，叠加 IP 限速 + (IP, echoID) 去重窗口；窗口内重复请求按幂等返回成功形状。
	likeMW := huma.Middlewares{humares.Bridge(middleware.RateLimitWithIdempotency(2, 5, time.Hour, "id", func(c *gin.Context) {
		c.JSON(http.StatusOK, commonModel.OK[any](nil, commonModel.LIKE_ECHO_SUCCESS))
	}))}
	reg(api, huma.Operation{
		OperationID: "echo-like",
		Method:      http.MethodPut,
		Path:        "/echo/like/{id}",
		Summary:     "点赞 Echo",
		Tags:        []string{"Echo"},
		Middlewares: likeMW,
	}, h.EchoHandler.LikeEcho)

	reg(api, huma.Operation{
		OperationID: "tag-list",
		Method:      http.MethodGet,
		Path:        "/tags",
		Summary:     "获取所有标签",
		Tags:        []string{"Tag"},
	}, h.EchoHandler.GetAllTags)

	// 可匿名降级读接口
	optRead := func(id, method, path, summary string) huma.Operation {
		return huma.Operation{OperationID: id, Method: method, Path: path, Summary: summary, Tags: []string{"Echo"}, Middlewares: optional}
	}
	reg(api, optRead("echo-query", http.MethodPost, "/echo/query", "统一查询 Echo 列表"), h.EchoHandler.QueryEchos)
	reg(api, optRead("echo-page-get", http.MethodGet, "/echo/page", "分页获取 Echo（Deprecated）"), h.EchoHandler.GetEchosByPageGet)
	reg(api, optRead("echo-page-post", http.MethodPost, "/echo/page", "分页获取 Echo（Deprecated）"), h.EchoHandler.GetEchosByPagePost)
	reg(api, optRead("echo-by-tag", http.MethodGet, "/echo/tag/{tagid}", "按标签获取 Echo（Deprecated）"), h.EchoHandler.GetEchosByTagId)
	reg(api, optRead("echo-today", http.MethodGet, "/echo/today", "获取今天的 Echo"), h.EchoHandler.GetTodayEchos)
	reg(api, optRead("echo-hot", http.MethodGet, "/echo/hot", "获取热门 Echo"), h.EchoHandler.GetHotEchos)
	reg(api, optRead("echo-random", http.MethodGet, "/echo/random", "随机返回一篇 Echo"), h.EchoHandler.GetRandomEcho)
	reg(api, optRead("echo-onthisday", http.MethodGet, "/echo/onthisday", "那年今日"), h.EchoHandler.GetOnThisDayEchos)
	reg(api, optRead("echo-get", http.MethodGet, "/echo/{id}", "获取指定 ID 的 Echo"), h.EchoHandler.GetEchoById)

	// 写接口（echo:write）
	writeOp := func(id, method, path, summary string) huma.Operation {
		return huma.Operation{OperationID: id, Method: method, Path: path, Summary: summary, Tags: []string{"Echo"}, Security: writeSec, Middlewares: write}
	}
	reg(api, writeOp("echo-create", http.MethodPost, "/echo", "创建新的 Echo"), h.EchoHandler.PostEcho)
	reg(api, writeOp("echo-update", http.MethodPut, "/echo", "更新 Echo"), h.EchoHandler.UpdateEcho)
	reg(api, writeOp("echo-delete", http.MethodDelete, "/echo/{id}", "删除 Echo"), h.EchoHandler.DeleteEcho)
	reg(api, writeOp("tag-create", http.MethodPost, "/tag", "创建标签"), h.EchoHandler.CreateTag)
	reg(api, writeOp("tag-delete", http.MethodDelete, "/tag/{id}", "删除标签"), h.EchoHandler.DeleteTag)
}
