package di

import (
	"context"
	"log"
	"strings"

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
	"github.com/lin-snow/ech0/internal/storage"
	"github.com/lin-snow/ech0/internal/task"
	"github.com/lin-snow/ech0/internal/transaction"

	virefs "github.com/lin-snow/VireFS"
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

// ProvideVireFS builds a virefs.FS based on the storage config.
func ProvideVireFS() virefs.FS {
	cfg := config.Config().Storage
	switch strings.ToLower(strings.TrimSpace(cfg.Mode)) {
	case "s3":
		return buildS3FS(cfg)
	default:
		return buildLocalFS(cfg)
	}
}

// ProvideURLResolver builds a URLResolver based on the storage config.
func ProvideURLResolver() storage.URLResolver {
	cfg := config.Config().Storage
	switch strings.ToLower(strings.TrimSpace(cfg.Mode)) {
	case "s3":
		return buildS3URLResolver(cfg)
	default:
		return buildLocalURLResolver()
	}
}

func buildLocalFS(cfg config.StorageConfig) virefs.FS {
	root := cfg.DataRoot
	if root == "" {
		root = "data/files"
	}
	fs, err := virefs.NewLocalFS(root,
		virefs.WithCreateRoot(),
		virefs.WithAtomicWrite(),
	)
	if err != nil {
		log.Printf("[storage] failed to create local FS: %v, falling back to defaults", err)
		fs, _ = virefs.NewLocalFS("data/files",
			virefs.WithCreateRoot(),
			virefs.WithAtomicWrite(),
		)
	}
	return fs
}

func buildLocalURLResolver() storage.URLResolver {
	return func(key string) string {
		return "/api/files/" + key
	}
}

func buildS3FS(cfg config.StorageConfig) virefs.FS {
	provider := mapProvider(cfg.Provider)

	var opts []virefs.ObjectOption
	if cfg.PathPrefix != "" {
		opts = append(opts, virefs.WithPrefix(strings.Trim(cfg.PathPrefix, "/")+"/"))
	}

	endpoint := normalizeEndpoint(cfg.Endpoint, cfg.UseSSL)

	fs, err := virefs.NewObjectFSFromConfig(context.Background(), &virefs.S3Config{
		Provider:  provider,
		Endpoint:  endpoint,
		Region:    cfg.Region,
		Bucket:    cfg.BucketName,
		AccessKey: cfg.AccessKey,
		SecretKey: cfg.SecretKey,
	}, opts...)
	if err != nil {
		log.Printf("[storage] failed to create S3 FS: %v, falling back to local", err)
		return buildLocalFS(cfg)
	}
	return fs
}

func buildS3URLResolver(cfg config.StorageConfig) storage.URLResolver {
	prefix := ""
	if cfg.PathPrefix != "" {
		prefix = strings.Trim(cfg.PathPrefix, "/") + "/"
	}

	cdnURL := strings.TrimSpace(cfg.CDNURL)
	if cdnURL != "" {
		if !strings.HasPrefix(strings.ToLower(cdnURL), "http://") &&
			!strings.HasPrefix(strings.ToLower(cdnURL), "https://") {
			protocol := "http"
			if cfg.UseSSL {
				protocol = "https"
			}
			cdnURL = protocol + "://" + cdnURL
		}
		cdnURL = strings.TrimRight(cdnURL, "/")
		return func(key string) string {
			return cdnURL + "/" + prefix + key
		}
	}

	endpoint := normalizeEndpoint(cfg.Endpoint, cfg.UseSSL)
	baseURL := strings.TrimRight(endpoint, "/") + "/" + cfg.BucketName
	return func(key string) string {
		return baseURL + "/" + prefix + key
	}
}

func normalizeEndpoint(endpoint string, useSSL bool) string {
	if endpoint == "" {
		return endpoint
	}
	lower := strings.ToLower(endpoint)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return endpoint
	}
	if useSSL {
		return "https://" + endpoint
	}
	return "http://" + endpoint
}

func mapProvider(raw string) virefs.Provider {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "minio":
		return virefs.ProviderMinIO
	case "r2":
		return virefs.ProviderR2
	default:
		return virefs.ProviderAWS
	}
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
