//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	"github.com/lin-snow/ech0/internal/app"
	"github.com/lin-snow/ech0/internal/cache"
	"github.com/lin-snow/ech0/internal/database"
	"github.com/lin-snow/ech0/internal/event"
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/metric"
	"github.com/lin-snow/ech0/internal/monitor"
	"github.com/lin-snow/ech0/internal/repository"
	runtimeCache "github.com/lin-snow/ech0/internal/runtime/cache"
	runtimeEvent "github.com/lin-snow/ech0/internal/runtime/event"
	runtimeHTTP "github.com/lin-snow/ech0/internal/runtime/http"
	runtimeTask "github.com/lin-snow/ech0/internal/runtime/task"
	"github.com/lin-snow/ech0/internal/service"
	"github.com/lin-snow/ech0/internal/storage"
	"github.com/lin-snow/ech0/internal/task"
	"github.com/lin-snow/ech0/internal/transaction"
	"gorm.io/gorm"
)

var AppSet = app.ProviderSet

var DomainSet = wire.NewSet(
	BuildHandlers,
	BuildTasker,
	BuildEventRegistrar,
)

var InfraSet = wire.NewSet(
	database.ProviderSet,
	event.ProviderSet,
	cache.ProviderSet,
	transaction.ProviderSet,
)

var RuntimeSet = wire.NewSet(
	runtimeHTTP.ProviderSet,
	runtimeEvent.ProviderSet,
	runtimeTask.ProviderSet,
	runtimeCache.ProviderSet,
)

var EventGraphSet = wire.NewSet(
	repository.EchoSet,
	service.EchoSet,

	repository.UserSet,
	service.UserSet,

	repository.TodoSet,
	service.TodoSet,

	repository.InboxSet,
	service.InboxSet,

	repository.KeyValueSet,
	repository.QueueSet,
	repository.WebhookSet,

	event.NewWebhookDispatcher,
	event.NewBackupScheduler,
	event.NewDeadLetterResolver,
	event.NewAgentProcessor,
	event.NewInboxDispatcher,
	event.NewEventHandlers,
	event.NewEventRegistry,
)

var HandlerGraphSet = wire.NewSet(
	storage.ProviderSet,
	repository.FileSet,
	handler.WebSet,

	repository.UserSet,
	service.UserSet,
	handler.UserSet,

	repository.EchoSet,
	service.EchoSet,
	handler.EchoSet,

	repository.CommonSet,
	service.CommonSet,
	handler.CommonSet,

	repository.WebhookSet,
	repository.KeyValueSet,

	repository.SettingSet,
	service.SettingSet,
	handler.SettingSet,

	repository.InboxSet,
	service.InboxSet,
	handler.InboxSet,

	repository.TodoSet,
	service.TodoSet,
	handler.TodoSet,

	repository.ConnectSet,
	service.ConnectSet,
	handler.ConnectSet,

	metric.NewSystemCollector,
	monitor.NewMonitor,

	service.DashboardSet,
	handler.DashboardSet,

	service.AgentSet,
	handler.AgentSet,

	service.BackupSet,
	handler.BackupSet,

	handler.NewBundle,
)

var TaskerGraphSet = wire.NewSet(
	storage.ProviderSet,
	repository.FileSet,
	repository.KeyValueSet,
	repository.WebhookSet,

	repository.SettingSet,
	service.SettingSet,

	repository.EchoSet,
	service.EchoSet,

	repository.CommonSet,
	service.CommonSet,

	repository.QueueSet,
	task.NewTasker,
)

// BuildWebApp 构建 Web 生命周期应用。
func BuildWebApp() (*app.App, func(), error) {
	wire.Build(
		InfraSet,
		DomainSet,
		RuntimeSet,
		AppSet,
	)
	return &app.App{}, nil, nil
}

// BuildApp 兼容旧入口，委托给 BuildWebApp。
func BuildApp() (*app.App, func(), error) {
	return BuildWebApp()
}

func BuildEventRegistrar(
	dbProvider func() *gorm.DB,
	ebProvider func() event.IEventBus,
	appCache cache.ICache[string, any],
	tx transaction.Transactor,
) (*event.EventRegistrar, error) {
	wire.Build(EventGraphSet)
	return &event.EventRegistrar{}, nil
}

// BuildHandlers 使用 wire 生成的代码来构建 Handlers 实例。
func BuildHandlers(
	dbProvider func() *gorm.DB,
	appCache cache.ICache[string, any],
	tx transaction.Transactor,
	ebProvider func() event.IEventBus,
) (*handler.Bundle, error) {
	wire.Build(HandlerGraphSet)
	return &handler.Bundle{}, nil
}

// BuildWebRuntime 构建 HTTP runtime（用于测试和独立启动场景）。
func BuildWebRuntime() (*runtimeHTTP.Runtime, error) {
	wire.Build(
		InfraSet,
		BuildHandlers,
		runtimeHTTP.ProviderSet,
	)
	return &runtimeHTTP.Runtime{}, nil
}

func BuildTasker(
	dbProvider func() *gorm.DB,
	appCache cache.ICache[string, any],
	tx transaction.Transactor,
	ebProvider func() event.IEventBus,
) (*task.Tasker, error) {
	wire.Build(TaskerGraphSet)
	return &task.Tasker{}, nil
}
