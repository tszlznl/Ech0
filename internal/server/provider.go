// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package server

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/lin-snow/ech0/internal/config"
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/middleware"
	"github.com/lin-snow/ech0/internal/router"
)

func ProvideGinEngine() *gin.Engine {
	if config.Config().Server.Mode == "debug" {
		gin.SetMode(gin.DebugMode)
		// Air 把二进制 stdout 接成非 TTY 管道，gin 默认关色；强制开彩色访问日志。
		gin.ForceConsoleColor()
		// 抑制启动时冗长的 [GIN-debug] 路由注册刷屏，保持 dev 终端干净。
		gin.DebugPrintRouteFunc = func(_, _, _ string, _ int) {}
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	return gin.New()
}

func ProvideHTTPServer(engine *gin.Engine, handlers *handler.Bundle, mwDeps *middleware.Deps) *Server {
	router.SetupRouter(engine, handlers, mwDeps)
	return New(engine)
}

var ProviderSet = wire.NewSet(ProvideGinEngine, ProvideHTTPServer)
