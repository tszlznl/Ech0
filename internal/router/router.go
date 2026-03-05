package router

import (
	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/di"
	"github.com/lin-snow/ech0/internal/middleware"
)

type AppRouterGroup struct {
	ResourceGroup     *gin.RouterGroup
	PublicRouterGroup *gin.RouterGroup
	AuthRouterGroup   *gin.RouterGroup
	WSRouterGroup     *gin.RouterGroup
}

// SetupRouter 配置路由
func SetupRouter(r *gin.Engine, h *di.Handlers) {
	// === 使用本地目录提供前端 ===)
	// // Setup Frontend
	// r.Use(static.Serve("/", static.LocalFile("./template", false)))
	// // 由于Vue3 和SPA模式，所以处理匹配不到的路由(重定向到index.html)
	// r.NoRoute(func(c *gin.Context) {
	// 	c.File("./template/index.html")
	// })

	ctx := &RouterContext{
		Engine:   r,
		Handlers: h,
	}

	for _, module := range coreRouteModules() {
		module.Register(ctx)
	}
	for _, module := range businessRouteModules() {
		module.Register(ctx)
	}
}

// setupRouterGroup 初始化路由组
func setupRouterGroup(r *gin.Engine) *AppRouterGroup {
	resource := r.Group("/")
	public := r.Group("/api")
	auth := r.Group("/api")
	auth.Use(middleware.NoCache(), middleware.JWTAuthMiddleware())
	ws := r.Group("/ws")
	return &AppRouterGroup{
		ResourceGroup:     resource,
		PublicRouterGroup: public,
		AuthRouterGroup:   auth,
		WSRouterGroup:     ws,
	}
}
