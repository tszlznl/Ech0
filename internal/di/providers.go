package di

import (
	"github.com/gin-gonic/gin"
	virefs "github.com/lin-snow/VireFS"
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
	"github.com/lin-snow/ech0/internal/storage"
	"github.com/lin-snow/ech0/internal/task"
	"github.com/lin-snow/ech0/internal/transaction"
	"gorm.io/gorm"
)

func ProvideCache(factory *cache.CacheFactory) cache.ICache[string, any] {
	return factory.Cache()
}

func ProvideCacheCleanup(factory *cache.CacheFactory) func() error {
	return factory.Cleanup
}

func ProvideTransactionManager(
	factory *transaction.TransactionManagerFactory,
) transaction.TransactionManager {
	return factory.TransactionManager()
}

func ProvideDBProvider() func() *gorm.DB {
	database.InitDatabase()
	return database.GetDB
}

func ProvideEventBusProvider() func() event.IEventBus {
	event.InitEventBus()
	return event.GetEventBus
}

func ProvideCacheFactory() *cache.CacheFactory {
	return cache.NewCacheFactory()
}

func ProvideVireFS() virefs.FS {
	return storage.NewFS(config.Config().Storage)
}

func ProvideURLResolver() storage.URLResolver {
	return storage.NewURLResolver(config.Config().Storage)
}

func ProvideTransactionManagerFactory(
	dbProvider func() *gorm.DB,
) *transaction.TransactionManagerFactory {
	return transaction.NewTransactionManagerFactory(dbProvider)
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

func ProvideHTTPRuntime(s *server.Server) *runtimeHTTP.Runtime {
	return runtimeHTTP.New(s)
}

func ProvideEventRuntime(registrar *event.EventRegistrar) *runtimeEvent.Runtime {
	return runtimeEvent.New(registrar)
}

func ProvideTaskRuntime(tasker *task.Tasker) *runtimeTask.Runtime {
	return runtimeTask.New(tasker)
}

func ProvideCacheRuntime(cleanup func() error) *runtimeCache.Runtime {
	return runtimeCache.New(cleanup)
}

func ProvideHandlers(
	dbProvider func() *gorm.DB,
	cacheFactory *cache.CacheFactory,
	tmFactory *transaction.TransactionManagerFactory,
	ebProvider func() event.IEventBus,
) (*handler.Bundle, error) {
	return BuildHandlers(dbProvider, cacheFactory, tmFactory, ebProvider)
}

func ProvideTasker(
	dbProvider func() *gorm.DB,
	cacheFactory *cache.CacheFactory,
	tmFactory *transaction.TransactionManagerFactory,
	ebProvider func() event.IEventBus,
) (*task.Tasker, error) {
	return BuildTasker(dbProvider, cacheFactory, tmFactory, ebProvider)
}

func ProvideEventRegistrar(
	dbProvider func() *gorm.DB,
	ebProvider func() event.IEventBus,
	cacheFactory *cache.CacheFactory,
	tmFactory *transaction.TransactionManagerFactory,
) (*event.EventRegistrar, error) {
	return BuildEventRegistrar(dbProvider, ebProvider, cacheFactory, tmFactory)
}

func ProvideWebComponents(
	cacheRuntime *runtimeCache.Runtime,
	eventRuntime *runtimeEvent.Runtime,
	taskRuntime *runtimeTask.Runtime,
	httpRuntime *runtimeHTTP.Runtime,
) []app.Component {
	return []app.Component{cacheRuntime, eventRuntime, taskRuntime, httpRuntime}
}

func ProvideApp(
	webComponents []app.Component,
) *app.App {
	return app.NewApp(webComponents)
}
