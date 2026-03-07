//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	"github.com/lin-snow/ech0/internal/app"
	"github.com/lin-snow/ech0/internal/cache"
	"github.com/lin-snow/ech0/internal/event"
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/metric"
	"github.com/lin-snow/ech0/internal/monitor"
	repository "github.com/lin-snow/ech0/internal/repository"
	runtimeCache "github.com/lin-snow/ech0/internal/runtime/cache"
	runtimeEvent "github.com/lin-snow/ech0/internal/runtime/event"
	runtimeHTTP "github.com/lin-snow/ech0/internal/runtime/http"
	runtimeTask "github.com/lin-snow/ech0/internal/runtime/task"
	service "github.com/lin-snow/ech0/internal/service"
	"github.com/lin-snow/ech0/internal/storage"
	"github.com/lin-snow/ech0/internal/task"
	"github.com/lin-snow/ech0/internal/transaction"
	"gorm.io/gorm"
)

var AppSet = wire.NewSet(
	ProvideWebComponents,
	app.ProviderSet,
)

var DomainSet = wire.NewSet(
	ProvideHandlers,
	ProvideTasker,
	ProvideEventRegistrar,
)

var InfraSet = wire.NewSet(
	ProvideDBProvider,
	ProvideEventBusProvider,
	cache.ProviderSet,
	transaction.ProviderSet,
	ProvideGinEngine,
)

var RuntimeSet = wire.NewSet(
	ProvideHTTPServer,
	runtimeHTTP.ProviderSet,
	runtimeEvent.ProviderSet,
	runtimeTask.ProviderSet,
	runtimeCache.ProviderSet,
)

var WebSet = wire.NewSet(handler.WebSet)
var UserSet = wire.NewSet(repository.UserSet, service.UserSet, handler.UserSet)
var EchoSet = wire.NewSet(repository.EchoSet, service.EchoSet, handler.EchoSet)
var CommonSet = wire.NewSet(repository.CommonSet, service.CommonSet, handler.CommonSet)
var KeyValueSet = wire.NewSet(repository.KeyValueSet)
var SettingSet = wire.NewSet(repository.SettingSet, service.SettingSet, handler.SettingSet)
var TodoSet = wire.NewSet(repository.TodoSet, service.TodoSet, handler.TodoSet)
var ConnectSet = wire.NewSet(repository.ConnectSet, service.ConnectSet, handler.ConnectSet)
var BackupSet = wire.NewSet(service.BackupSet, handler.BackupSet)
var DashboardSet = wire.NewSet(service.DashboardSet, handler.DashboardSet)
var AgentSet = wire.NewSet(service.AgentSet, handler.AgentSet)
var WebhookSet = wire.NewSet(repository.WebhookSet)
var InboxSet = wire.NewSet(repository.InboxSet, service.InboxSet, handler.InboxSet)
var QueueSet = wire.NewSet(repository.QueueSet)

var TaskSet = wire.NewSet(task.NewTasker)
var EventSet = wire.NewSet(
	event.NewWebhookDispatcher,
	event.NewBackupScheduler,
	event.NewDeadLetterResolver,
	event.NewAgentProcessor,
	event.NewInboxDispatcher,
	event.NewEventHandlers,
	event.NewEventRegistry,
)
var MetricSet = wire.NewSet(metric.NewSystemCollector)
var MonitorSet = wire.NewSet(monitor.NewMonitor)

// BuildApp 构建应用内核。
func BuildApp() (*app.App, func(), error) {
	wire.Build(
		InfraSet,
		DomainSet,
		RuntimeSet,
		AppSet,
	)
	return &app.App{}, nil, nil
}

func BuildEventRegistrar(
	dbProvider func() *gorm.DB,
	ebProvider func() event.IEventBus,
	cacheFactory *cache.CacheFactory,
	tmFactory *transaction.TransactionManagerFactory,
) (*event.EventRegistrar, error) {
	wire.Build(
		EchoSet,
		UserSet,
		TodoSet,
		InboxSet,
		cache.CacheSet,
		transaction.ManagerSet,
		KeyValueSet,
		QueueSet,
		WebhookSet,
		EventSet,
	)
	return &event.EventRegistrar{}, nil
}

// BuildHandlers 使用 wire 生成的代码来构建 Handlers 实例。
func BuildHandlers(
	dbProvider func() *gorm.DB,
	cacheFactory *cache.CacheFactory,
	tmFactory *transaction.TransactionManagerFactory,
	ebProvider func() event.IEventBus,
) (*handler.Bundle, error) {
	wire.Build(
		cache.CacheSet,
		storage.ProviderSet,
		repository.FileSet,
		transaction.ManagerSet,
		WebSet,
		UserSet,
		EchoSet,
		CommonSet,
		WebhookSet,
		KeyValueSet,
		SettingSet,
		InboxSet,
		TodoSet,
		ConnectSet,
		MetricSet,
		MonitorSet,
		DashboardSet,
		AgentSet,
		BackupSet,
		handler.NewBundle,
	)
	return &handler.Bundle{}, nil
}

// BuildWebRuntime 构建 HTTP runtime（用于测试和独立启动场景）。
func BuildWebRuntime() (*runtimeHTTP.Runtime, error) {
	wire.Build(
		InfraSet,
		DomainSet,
		ProvideHTTPServer,
		runtimeHTTP.ProviderSet,
	)
	return &runtimeHTTP.Runtime{}, nil
}

func BuildTasker(
	dbProvider func() *gorm.DB,
	cacheFactory *cache.CacheFactory,
	tmFactory *transaction.TransactionManagerFactory,
	ebProvider func() event.IEventBus,
) (*task.Tasker, error) {
	wire.Build(
		cache.CacheSet,
		storage.ProviderSet,
		repository.FileSet,
		KeyValueSet,
		transaction.ManagerSet,
		WebhookSet,
		SettingSet,
		EchoSet,
		CommonSet,
		QueueSet,
		TaskSet,
	)
	return &task.Tasker{}, nil
}
