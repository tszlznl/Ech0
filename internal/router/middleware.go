// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/config"
	i18nUtil "github.com/lin-snow/ech0/internal/i18n"
	"github.com/lin-snow/ech0/internal/middleware"
)

// setupMiddleware 设置中间件
func setupMiddleware(r *gin.Engine) {
	// Dev-only 彩色访问日志（绿色状态码 / 彩色方法徽章）。放在 Recovery 之外（更外层），
	// 这样被 Recovery 兜住的 panic 仍能打出带最终 500 状态的访问行。Prod（release）不挂。
	if config.Config().Server.Mode == "debug" {
		r.Use(gin.Logger())
	}
	// Recovery middleware to recover from any panics and write a 500 if there was one.
	r.Use(gin.Recovery())
	// Powered-by header
	r.Use(middleware.PoweredBy())
	// Cors middleware
	r.Use(middleware.Cors())
	// Locale and request localizer middleware
	r.Use(i18nUtil.Middleware())
	// Global write guard middleware
	r.Use(middleware.WriteGuard())
}
