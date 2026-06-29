// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package handler 暴露仪表盘相关的 HTTP 接口。
//
// JSON 端点（检查更新/历史日志/访客统计）走 Huma type-first；
// 日志的 SSE / WebSocket 实时订阅仍走裸 gin（见本文件下方 + setupDashboardRoutes）。
package handler

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/handler/humares"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	service "github.com/lin-snow/ech0/internal/service/dashboard"
	githubUtil "github.com/lin-snow/ech0/internal/util/github"
	jwtUtil "github.com/lin-snow/ech0/internal/util/jwt"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	versionPkg "github.com/lin-snow/ech0/internal/version"
	"github.com/lin-snow/ech0/internal/visitor"
	"go.uber.org/zap"
	"golang.org/x/mod/semver"
)

type DashboardHandler struct {
	dashboardService service.Service
}

func NewDashboardHandler(dashboardService service.Service) *DashboardHandler {
	return &DashboardHandler{
		dashboardService: dashboardService,
	}
}

type (
	CheckUpdateInput   struct{}
	GetSystemLogsInput struct {
		Tail    string `query:"tail" doc:"返回最近多少条（默认 200）"`
		Level   string `query:"level" doc:"日志级别过滤"`
		Keyword string `query:"keyword" doc:"关键词过滤"`
	}
	GetVisitorStatsInput struct{}

	// CheckUpdateResponse 版本检查结果。
	CheckUpdateResponse struct {
		CurrentVersion string `json:"current_version"`
		LatestVersion  string `json:"latest_version"`
		HasUpdate      bool   `json:"has_update"`
	}
)

// CheckUpdate 检查 Ech0 版本更新（admin:settings）。
func (dashboardHandler *DashboardHandler) CheckUpdate(ctx context.Context, _ *CheckUpdateInput) (*humares.Envelope[CheckUpdateResponse], error) {
	latestVersion, err := githubUtil.GetLatestVersion()
	if err != nil {
		return nil, humares.Err(ctx, err)
	}

	cur := semver.Canonical("v" + strings.TrimPrefix(strings.TrimSpace(versionPkg.Version), "v"))
	lat := semver.Canonical("v" + strings.TrimPrefix(strings.TrimSpace(latestVersion), "v"))
	hasUpdate := cur != "" && lat != "" && semver.Compare(lat, cur) > 0

	return humares.OK(ctx, CheckUpdateResponse{
		CurrentVersion: versionPkg.Version,
		LatestVersion:  latestVersion,
		HasUpdate:      hasUpdate,
	}), nil
}

// GetSystemLogs 获取系统历史日志（admin:settings）。
func (dashboardHandler *DashboardHandler) GetSystemLogs(ctx context.Context, in *GetSystemLogsInput) (*humares.Envelope[[]logUtil.LogEntry], error) {
	tail := 200
	if rawTail := strings.TrimSpace(in.Tail); rawTail != "" {
		n, err := strconv.Atoi(rawTail)
		if err != nil || n <= 0 {
			return nil, humares.Err(ctx, commonModel.NewBizErrorWithMessageKey(
				commonModel.ErrCodeInvalidQuery, commonModel.INVALID_QUERY_PARAMS, commonModel.MsgKeyDashboardTailBad, nil,
			))
		}
		tail = n
	}

	logs, err := dashboardHandler.dashboardService.GetSystemLogs(service.SystemLogQuery{
		Tail:    tail,
		Level:   in.Level,
		Keyword: in.Keyword,
	})
	if err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OKKeyed(ctx, logs, "获取系统日志成功", commonModel.MsgKeyDashboardLogsOk, nil), nil
}

// GetVisitorStats 获取近七天访客统计（admin:settings）。
func (dashboardHandler *DashboardHandler) GetVisitorStats(ctx context.Context, _ *GetVisitorStatsInput) (*humares.Envelope[[]visitor.DayStat], error) {
	return humares.OK(ctx, dashboardHandler.dashboardService.GetVisitorStats()), nil
}

// --- 以下为实时日志订阅，仍走裸 gin（WebSocket / SSE） ---

// WSSubscribeSystemLogs 通过 WebSocket 订阅系统日志。
func (dashboardHandler *DashboardHandler) WSSubscribeSystemLogs() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.Query("token")
		if token == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "missing token"})
			return
		}

		token = strings.Trim(token, `"`)
		if _, err := jwtUtil.ParseToken(token); err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "invalid token"})
			return
		}

		err := dashboardHandler.dashboardService.WSSubscribeSystemLogs(
			ctx.Writer,
			ctx.Request,
			service.SystemLogStreamFilter{
				Level:   ctx.Query("level"),
				Keyword: ctx.Query("keyword"),
			},
		)
		if err != nil {
			logUtil.GetLogger().Error("WebSocket Subscribe System Logs Failed", zap.Error(err))
		}
	}
}

// SSESubscribeSystemLogs 通过 SSE 订阅系统日志（WS 兜底）。
func (dashboardHandler *DashboardHandler) SSESubscribeSystemLogs() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.Query("token")
		if token == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "missing token"})
			return
		}

		token = strings.Trim(token, `"`)
		if _, err := jwtUtil.ParseToken(token); err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "invalid token"})
			return
		}

		err := dashboardHandler.dashboardService.SSESubscribeSystemLogs(
			ctx.Writer,
			ctx.Request,
			service.SystemLogStreamFilter{
				Level:   ctx.Query("level"),
				Keyword: ctx.Query("keyword"),
			},
		)
		if err != nil {
			logUtil.GetLogger().Error("SSE Subscribe System Logs Failed", zap.Error(err))
		}
	}
}
