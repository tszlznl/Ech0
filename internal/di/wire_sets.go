//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	"github.com/lin-snow/ech0/internal/cache"
	"github.com/lin-snow/ech0/internal/event"
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/metric"
	"github.com/lin-snow/ech0/internal/monitor"
	"github.com/lin-snow/ech0/internal/repository"
	"github.com/lin-snow/ech0/internal/service"
	"github.com/lin-snow/ech0/internal/storage"
	"github.com/lin-snow/ech0/internal/task"
	"github.com/lin-snow/ech0/internal/transaction"
)

var UserFeatureSet = wire.NewSet(repository.UserSet, service.UserSet, handler.UserSet)
var EchoFeatureSet = wire.NewSet(repository.EchoSet, service.EchoSet, handler.EchoSet)
var CommonFeatureSet = wire.NewSet(repository.CommonSet, service.CommonSet, handler.CommonSet)
var SettingFeatureSet = wire.NewSet(repository.SettingSet, service.SettingSet, handler.SettingSet)
var TodoFeatureSet = wire.NewSet(repository.TodoSet, service.TodoSet, handler.TodoSet)
var ConnectFeatureSet = wire.NewSet(repository.ConnectSet, service.ConnectSet, handler.ConnectSet)
var InboxFeatureSet = wire.NewSet(repository.InboxSet, service.InboxSet, handler.InboxSet)
var BackupFeatureSet = wire.NewSet(service.BackupSet, handler.BackupSet)
var DashboardFeatureSet = wire.NewSet(service.DashboardSet, handler.DashboardSet)
var AgentFeatureSet = wire.NewSet(service.AgentSet, handler.AgentSet)

var EventGraphSet = wire.NewSet(
	EchoFeatureSet,
	UserFeatureSet,
	TodoFeatureSet,
	InboxFeatureSet,
	cache.CacheSet,
	transaction.ManagerSet,
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
	cache.CacheSet,
	storage.ProviderSet,
	repository.FileSet,
	transaction.ManagerSet,
	handler.WebSet,
	UserFeatureSet,
	EchoFeatureSet,
	CommonFeatureSet,
	repository.WebhookSet,
	repository.KeyValueSet,
	SettingFeatureSet,
	InboxFeatureSet,
	TodoFeatureSet,
	ConnectFeatureSet,
	metric.NewSystemCollector,
	monitor.NewMonitor,
	DashboardFeatureSet,
	AgentFeatureSet,
	BackupFeatureSet,
	handler.NewBundle,
)

var TaskerGraphSet = wire.NewSet(
	cache.CacheSet,
	storage.ProviderSet,
	repository.FileSet,
	repository.KeyValueSet,
	transaction.ManagerSet,
	repository.WebhookSet,
	SettingFeatureSet,
	EchoFeatureSet,
	CommonFeatureSet,
	repository.QueueSet,
	task.NewTasker,
)
