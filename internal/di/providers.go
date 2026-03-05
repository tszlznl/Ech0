package di

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/app"
	"github.com/lin-snow/ech0/internal/cache"
	"github.com/lin-snow/ech0/internal/config"
	"github.com/lin-snow/ech0/internal/database"
	"github.com/lin-snow/ech0/internal/event"
	"github.com/lin-snow/ech0/internal/handler"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	keyvalueRepository "github.com/lin-snow/ech0/internal/repository/keyvalue"
	"github.com/lin-snow/ech0/internal/router"
	runtimeCache "github.com/lin-snow/ech0/internal/runtime/cache"
	runtimeEvent "github.com/lin-snow/ech0/internal/runtime/event"
	runtimeHTTP "github.com/lin-snow/ech0/internal/runtime/http"
	runtimeTask "github.com/lin-snow/ech0/internal/runtime/task"
	"github.com/lin-snow/ech0/internal/server"
	"github.com/lin-snow/ech0/internal/storage"
	storageFactory "github.com/lin-snow/ech0/internal/storage/factory"
	"github.com/lin-snow/ech0/internal/task"
	"github.com/lin-snow/ech0/internal/transaction"
	httpUtil "github.com/lin-snow/ech0/internal/util/http"
	jsonUtil "github.com/lin-snow/ech0/internal/util/json"
	"github.com/spf13/afero"
	"gorm.io/gorm"
)

// ProvideCache 提供通用缓存实例给 wire 注入
func ProvideCache(factory *cache.CacheFactory) cache.ICache[string, any] {
	return factory.Cache()
}

// ProvideCacheCleanup 提供缓存清理函数给生命周期管理使用
func ProvideCacheCleanup(factory *cache.CacheFactory) func() error {
	return factory.Cleanup
}

// ProvideTransactionManager 提供事务管理器实例给 wire 注入
func ProvideTransactionManager(
	factory *transaction.TransactionManagerFactory,
) transaction.TransactionManager {
	return factory.TransactionManager()
}

// ProvideDBProvider 提供数据库 Provider。
func ProvideDBProvider() func() *gorm.DB {
	database.InitDatabase()
	return database.GetDB
}

// ProvideEventBusProvider 提供 EventBus Provider。
func ProvideEventBusProvider() func() event.IEventBus {
	event.InitEventBus()
	return event.GetEventBus
}

// ProvideCacheFactory 创建缓存工厂。
func ProvideCacheFactory() *cache.CacheFactory {
	return cache.NewCacheFactory()
}

func ProvideAferoFs() afero.Fs {
	return afero.NewOsFs()
}

func ProvideStoragePort(
	fs afero.Fs,
	keyvalueRepo keyvalueRepository.KeyValueRepositoryInterface,
) storage.StoragePort {
	mode := storageFactory.ModeLocal
	var s3Setting settingModel.S3Setting

	if value, err := keyvalueRepo.GetKeyValue(commonModel.S3SettingKey); err == nil && strings.TrimSpace(value) != "" {
		if err := jsonUtil.JSONUnmarshal([]byte(value), &s3Setting); err == nil && s3Setting.Enable {
			s3Setting.Endpoint = httpUtil.TrimURL(s3Setting.Endpoint)
			mode = storageFactory.ModeS3
		}
	}

	port, err := storageFactory.Build(storageFactory.BuildInput{
		Mode:      mode,
		FS:        fs,
		S3Setting: s3Setting,
	})
	if err != nil {
		port, _ = storageFactory.Build(storageFactory.BuildInput{Mode: storageFactory.ModeLocal, FS: fs})
	}
	return port
}

// ProvideTransactionManagerFactory 创建事务管理器工厂。
func ProvideTransactionManagerFactory(
	dbProvider func() *gorm.DB,
) *transaction.TransactionManagerFactory {
	return transaction.NewTransactionManagerFactory(dbProvider)
}

// ProvideGinEngine 创建 Gin 引擎。
func ProvideGinEngine() *gin.Engine {
	if config.Config().Server.Mode == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	return gin.New()
}

// ProvideHTTPServer 创建并装配纯 HTTP runtime。
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

// ProvideWebComponents 组装 Web 组件启动顺序。
// 启动顺序: cache(no-op) -> event -> task -> http
// 停止顺序: http -> task -> event -> cache
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
