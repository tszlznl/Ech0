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
	jwtUtil "github.com/lin-snow/ech0/internal/util/jwt"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"go.uber.org/zap"
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
					Msg: commonModel.INVALID_QUERY_PARAMS,
					Err: err,
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
			Data: logs,
			Msg:  "获取系统日志成功",
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
