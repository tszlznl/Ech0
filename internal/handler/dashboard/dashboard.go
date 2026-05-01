// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	res "github.com/lin-snow/ech0/internal/handler/response"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	service "github.com/lin-snow/ech0/internal/service/dashboard"
	githubUtil "github.com/lin-snow/ech0/internal/util/github"
	jwtUtil "github.com/lin-snow/ech0/internal/util/jwt"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	versionPkg "github.com/lin-snow/ech0/internal/version"
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

// GetSystemLogs 获取系统历史日志
func (dashboardHandler *DashboardHandler) GetSystemLogs() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		tail := 200
		if rawTail := strings.TrimSpace(ctx.Query("tail")); rawTail != "" {
			n, err := strconv.Atoi(rawTail)
			if err != nil || n <= 0 {
				if err == nil {
					err = errors.New("tail must be greater than zero")
				}
				return res.Response{
					Msg:        commonModel.INVALID_QUERY_PARAMS,
					ErrorCode:  commonModel.ErrCodeInvalidQuery,
					MessageKey: commonModel.MsgKeyDashboardTailBad,
					Err:        err,
				}
			}
			tail = n
		}

		logs, err := dashboardHandler.dashboardService.GetSystemLogs(service.SystemLogQuery{
			Tail:    tail,
			Level:   ctx.Query("level"),
			Keyword: ctx.Query("keyword"),
		})
		if err != nil {
			return res.Response{Err: err}
		}
		return res.Response{
			Data:       logs,
			Msg:        "获取系统日志成功",
			MessageKey: commonModel.MsgKeyDashboardLogsOk,
		}
	})
}

// GetVisitorStats 获取近七天访客统计
func (dashboardHandler *DashboardHandler) GetVisitorStats() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		return res.Response{
			Data: dashboardHandler.dashboardService.GetVisitorStats(),
			Msg:  commonModel.SUCCESS_MESSAGE,
		}
	})
}

// WSSubscribeSystemLogs 通过 WebSocket 订阅系统日志
func (dashboardHandler *DashboardHandler) WSSubscribeSystemLogs() gin.HandlerFunc {
	return gin.HandlerFunc(func(ctx *gin.Context) {
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
	})
}

// CheckUpdate 检查 Ech0 版本更新
func (dashboardHandler *DashboardHandler) CheckUpdate() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		latestVersion, err := githubUtil.GetLatestVersion()
		if err != nil {
			return res.Response{Msg: "检查更新失败", Err: err}
		}

		cur := semver.Canonical("v" + strings.TrimPrefix(strings.TrimSpace(versionPkg.Version), "v"))
		lat := semver.Canonical("v" + strings.TrimPrefix(strings.TrimSpace(latestVersion), "v"))
		hasUpdate := cur != "" && lat != "" && semver.Compare(lat, cur) > 0

		return res.Response{
			Data: gin.H{
				"current_version": versionPkg.Version,
				"latest_version":  latestVersion,
				"has_update":      hasUpdate,
			},
			Msg: commonModel.SUCCESS_MESSAGE,
		}
	})
}

// SSESubscribeSystemLogs 通过 SSE 订阅系统日志（WS 兜底）。
func (dashboardHandler *DashboardHandler) SSESubscribeSystemLogs() gin.HandlerFunc {
	return gin.HandlerFunc(func(ctx *gin.Context) {
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
	})
}
