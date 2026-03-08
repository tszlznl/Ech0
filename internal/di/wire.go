//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	busen "github.com/lin-snow/Busen"
	"github.com/lin-snow/ech0/internal/app"
	"github.com/lin-snow/ech0/internal/cache"
	"github.com/lin-snow/ech0/internal/database"
	eventbus "github.com/lin-snow/ech0/internal/event/bus"
	eventpublisher "github.com/lin-snow/ech0/internal/event/publisher"
	eventregistry "github.com/lin-snow/ech0/internal/event/registry"
	eventsubscriber "github.com/lin-snow/ech0/internal/event/subscriber"
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/metric"
	"github.com/lin-snow/ech0/internal/monitor"
	"github.com/lin-snow/ech0/internal/repository"
	"github.com/lin-snow/ech0/internal/server"
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
	ProvideBackupScheduleApplier,
	BuildEventRegistrar,
	eventregistry.ProvideRegisteredRegistrar,
)

var InfraSet = wire.NewSet(
	database.ProviderSet,
	eventbus.ProvideProvider,
	cache.ProviderSet,
	transaction.ProviderSet,
)

var RuntimeSet = server.ProviderSet

var EventGraphSet = wire.NewSet(
	repository.EchoSet,

	repository.UserSet,

	repository.TodoSet,

	repository.InboxSet,

	repository.KeyValueSet,
	repository.QueueSet,
	repository.WebhookSet,

	wire.Bind(new(eventregistry.WebhookObserver), new(*eventsubscriber.WebhookDispatcher)),
	wire.Bind(new(eventsubscriber.DeadLetterProcessor), new(*eventsubscriber.WebhookDispatcher)),
	wire.Bind(new(eventregistry.DeadLetterHandler), new(*eventsubscriber.DeadLetterResolver)),
	wire.Bind(new(eventregistry.BackupScheduleHandler), new(*eventsubscriber.BackupScheduler)),
	wire.Bind(new(eventregistry.AgentEventHandler), new(*eventsubscriber.AgentProcessor)),
	wire.Bind(new(eventregistry.InboxEventHandler), new(*eventsubscriber.InboxDispatcher)),

	eventsubscriber.NewWebhookDispatcher,
	eventsubscriber.NewBackupScheduler,
	eventsubscriber.NewDeadLetterResolver,
	eventsubscriber.NewAgentProcessor,
	eventsubscriber.NewInboxDispatcher,
	eventregistry.NewEventHandlers,
	eventregistry.NewEventRegistry,
)

var HandlerGraphSet = wire.NewSet(
	eventpublisher.New,
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
	eventpublisher.New,
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
	task.ProviderSet,
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
	ebProvider func() *busen.Bus,
	appCache cache.ICache[string, any],
	tx transaction.Transactor,
	backupScheduleApplier eventsubscriber.BackupScheduleApplier,
) (*eventregistry.EventRegistrar, error) {
	wire.Build(EventGraphSet)
	return &eventregistry.EventRegistrar{}, nil
}

// BuildHandlers 使用 wire 生成的代码来构建 Handlers 实例。
func BuildHandlers(
	dbProvider func() *gorm.DB,
	appCache cache.ICache[string, any],
	tx transaction.Transactor,
	ebProvider func() *busen.Bus,
) (*handler.Bundle, error) {
	wire.Build(HandlerGraphSet)
	return &handler.Bundle{}, nil
}

// BuildServer 构建 HTTP server
func BuildServer() (*server.Server, error) {
	wire.Build(
		InfraSet,
		BuildHandlers,
		server.ProviderSet,
	)
	return &server.Server{}, nil
}

func BuildTasker(
	dbProvider func() *gorm.DB,
	appCache cache.ICache[string, any],
	tx transaction.Transactor,
	ebProvider func() *busen.Bus,
) (*task.Tasker, error) {
	wire.Build(TaskerGraphSet)
	return &task.Tasker{}, nil
}

func ProvideBackupScheduleApplier(t *task.Tasker) eventsubscriber.BackupScheduleApplier {
	return t
}
