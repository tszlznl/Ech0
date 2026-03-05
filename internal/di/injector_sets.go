//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	"github.com/lin-snow/ech0/internal/event"
	fediverse "github.com/lin-snow/ech0/internal/fediverse"
	handler "github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/metric"
	"github.com/lin-snow/ech0/internal/monitor"
	repository "github.com/lin-snow/ech0/internal/repository"
	service "github.com/lin-snow/ech0/internal/service"
	"github.com/lin-snow/ech0/internal/task"
)

var CacheSet = wire.NewSet(ProvideCache)
var TransactionManagerSet = wire.NewSet(ProvideTransactionManager)

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
var FediverseCoreSet = wire.NewSet(fediverse.NewFediverseCore)
var FediverseSet = wire.NewSet(
	repository.FediverseSet,
	service.FediverseSet,
	handler.FediverseSet,
	event.NewFediverseAgent,
)
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
