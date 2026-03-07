package di

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/app"
	"github.com/lin-snow/ech0/internal/cache"
	"github.com/lin-snow/ech0/internal/config"
	"github.com/lin-snow/ech0/internal/database"
	"github.com/lin-snow/ech0/internal/event"
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/router"
	runtimeCache "github.com/lin-snow/ech0/internal/runtime/cache"
	runtimeEvent "github.com/lin-snow/ech0/internal/runtime/event"
	runtimeHTTP "github.com/lin-snow/ech0/internal/runtime/http"
	runtimeTask "github.com/lin-snow/ech0/internal/runtime/task"
	"github.com/lin-snow/ech0/internal/server"
	"github.com/lin-snow/ech0/internal/task"
	"github.com/lin-snow/ech0/internal/transaction"
	"gorm.io/gorm"
)

func ProvideDBProvider() func() *gorm.DB {
	var once sync.Once
	return func() *gorm.DB {
		once.Do(database.InitDatabase)
		return database.GetDB()
	}
}

func ProvideEventBusProvider() func() event.IEventBus {
	var once sync.Once
	return func() event.IEventBus {
		once.Do(event.InitEventBus)
		return event.GetEventBus()
	}
}

func ProvideGinEngine() *gin.Engine {
	if config.Config().Server.Mode == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	return gin.New()
}

func ProvideHTTPServer(engine *gin.Engine, handlers *handler.Bundle) *server.Server {
	router.SetupRouter(engine, handlers)
	return server.New(engine)
}

func ProvideHandlers(
	dbProvider func() *gorm.DB,
	appCache cache.ICache[string, any],
	tx transaction.Transactor,
	ebProvider func() event.IEventBus,
) (*handler.Bundle, error) {
	return BuildHandlers(dbProvider, appCache, tx, ebProvider)
}

func ProvideTasker(
	dbProvider func() *gorm.DB,
	appCache cache.ICache[string, any],
	tx transaction.Transactor,
	ebProvider func() event.IEventBus,
) (*task.Tasker, error) {
	return BuildTasker(dbProvider, appCache, tx, ebProvider)
}

func ProvideEventRegistrar(
	dbProvider func() *gorm.DB,
	ebProvider func() event.IEventBus,
	appCache cache.ICache[string, any],
	tx transaction.Transactor,
) (*event.EventRegistrar, error) {
	return BuildEventRegistrar(dbProvider, ebProvider, appCache, tx)
}

func ProvideComponents(
	eventRuntime *runtimeEvent.Runtime,
	taskRuntime *runtimeTask.Runtime,
	httpRuntime *runtimeHTTP.Runtime,
) []app.Component {
	// 启动顺序必须保持为 event -> task -> http：
	// 1. 先注册事件处理器，避免后台任务启动后发布的事件没有订阅者。
	// 2. 再启动任务调度器，让任务依赖的事件流已经可用。
	// 3. 最后开放 HTTP 入口，避免服务对外可见时后台基础设施尚未完成初始化。
	return []app.Component{eventRuntime, taskRuntime, httpRuntime}
}

func ProvideShutdownHooks(cacheHook *runtimeCache.ShutdownHook) []app.ShutdownHook {
	return []app.ShutdownHook{cacheHook}
}
