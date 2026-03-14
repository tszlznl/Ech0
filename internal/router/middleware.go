package router

import (
	"github.com/gin-gonic/gin"
	i18nUtil "github.com/lin-snow/ech0/internal/i18n"
	"github.com/lin-snow/ech0/internal/middleware"
)

// setupMiddleware 设置中间件
func setupMiddleware(r *gin.Engine) {
	// Recovery middleware to recover from any panics and write a 500 if there was one.
	r.Use(gin.Recovery())
	// Cors middleware
	r.Use(middleware.Cors())
	// Locale and request localizer middleware
	r.Use(i18nUtil.Middleware())
	// Global write guard middleware
	r.Use(middleware.WriteGuard())
}
