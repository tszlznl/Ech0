// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	"github.com/lin-snow/ech0/internal/app"
	"github.com/lin-snow/ech0/internal/cache"
	"github.com/lin-snow/ech0/internal/database"
	eventbus "github.com/lin-snow/ech0/internal/event/bus"
	eventpublisher "github.com/lin-snow/ech0/internal/event/publisher"
	eventregistry "github.com/lin-snow/ech0/internal/event/registry"
	eventsubscriber "github.com/lin-snow/ech0/internal/event/subscriber"
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/job"
	jobRunner "github.com/lin-snow/ech0/internal/job/runner"
	"github.com/lin-snow/ech0/internal/middleware"
	"github.com/lin-snow/ech0/internal/migrator"
	jobModel "github.com/lin-snow/ech0/internal/model/job"
	"github.com/lin-snow/ech0/internal/repository"
	keyvalueRepository "github.com/lin-snow/ech0/internal/repository/keyvalue"
	"github.com/lin-snow/ech0/internal/server"
	"github.com/lin-snow/ech0/internal/service"
	commentService "github.com/lin-snow/ech0/internal/service/comment"
	copilotService "github.com/lin-snow/ech0/internal/service/copilot"
	migratorService "github.com/lin-snow/ech0/internal/service/migrator"
	userService "github.com/lin-snow/ech0/internal/service/user"
	"github.com/lin-snow/ech0/internal/storage"
	"github.com/lin-snow/ech0/internal/task"
	"github.com/lin-snow/ech0/internal/task/scheduled"
	"github.com/lin-snow/ech0/internal/transaction"
	"github.com/lin-snow/ech0/internal/visitor"
	"github.com/lin-snow/ech0/internal/webhook"
	"github.com/lin-snow/ech0/pkg/busen"
	"gorm.io/gorm"
)

var AppSet = app.ProviderSet

// VisitorSet 独立于 HandlerSet/TaskerSet,避免 wire 为两个 Build 各自生成一个 Tracker
// 导致"WebHandler 写入 #1、Tasker 从 #2 读出恒为 0"的 bug。必须在 BuildApp/BuildServer
// 顶层引入一次,统一下沉给 BuildHandlers 和 BuildTasker。
var VisitorSet = wire.NewSet(visitor.NewTracker)

// ProvideJobManager 构造已装配好 Runner 的共享单例 *job.Manager（在构造期一次性
// 完成注册）。Runner 只依赖 EmbeddingService / migrator.Importer（均不含 *job.Manager），
// 故不会与「MigratorService 需要 Manager」形成构造环。
func ProvideJobManager(
	repo job.JobRepository,
	reindex *jobRunner.ReindexRunner,
	migration *jobRunner.MigrationRunner,
) *job.Manager {
	m := job.NewManager(repo)
	m.Register(jobModel.TypeReindex, job.Adapt(reindex.Run))
	m.Register(jobModel.TypeMigration, job.Adapt(migration.Run))
	return m
}

// ProvideTaskManager 把各领域定时 Task 收进共享单例 *task.Manager（对应 ProvideJobManager）。
// NewManager 是变参，wire 无法直接喂，故在此把具体 Task 收口成一次构造。
func ProvideTaskManager(
	cleanup *scheduled.Cleanup,
	deadletter *scheduled.DeadLetter,
	backup *scheduled.Backup,
	visitorSnapshot *scheduled.VisitorSnapshot,
) (*task.Manager, error) {
	return task.NewManager(cleanup, deadletter, backup, visitorSnapshot)
}

// StorageSet 提供进程级共享单例 *storage.Manager。storage.Manager 是有状态基础设施
// （缓存当前存储后端，ReloadFromConfigAndDB 会改写它），必须全进程一个实例，否则
// 「设置页 / 迁移改了 S3 → 只 reload 了自己那份 Manager，文件服务仍用旧后端」。同
// VisitorSet：顶层引入一次，统一下沉给 BuildHandlers/BuildTasker/BuildJobManager。
// 它自带一份 KeyValueRepository 仅供读取 S3 设置，与各 Build 内的 KeyValueSet 互不冲突。
var StorageSet = wire.NewSet(
	keyvalueRepository.NewKeyValueRepository,
	wire.Bind(new(storage.S3SettingStore), new(*keyvalueRepository.KeyValueRepository)),
	storage.ProviderSet,
)

var DomainSet = wire.NewSet(
	BuildHandlers,
	BuildMiddlewares,
	BuildTasker,
	BuildMigrator,
	BuildJobManager,
	ProvideBackupScheduleApplier,
	BuildEventRegistrar,
)

var InfraSet = wire.NewSet(
	database.ProviderSet,
	eventbus.ProvideProvider,
	cache.ProviderSet,
	transaction.ProviderSet,
)

var RuntimeSet = server.ProviderSet

var EventSet = wire.NewSet(
	repository.EchoSet,

	repository.UserSet,

	repository.KeyValueSet,
	repository.QueueSet,
	repository.WebhookSet,
	repository.EmbeddingSet,

	wire.Bind(new(eventregistry.WebhookObserver), new(*webhook.Dispatcher)),
	wire.Bind(new(eventsubscriber.DeadLetterProcessor), new(*webhook.Dispatcher)),

	webhook.NewDispatcher,
	eventsubscriber.NewBackupScheduler,
	eventsubscriber.NewDeadLetterResolver,
	eventsubscriber.NewAgentProcessor,
	eventsubscriber.NewEmbeddingProcessor,
	service.EmbeddingSet,
	ProvideSubscriptionProviders,
	eventregistry.NewEventRegistry,
)

var HandlerSet = wire.NewSet(
	eventpublisher.New,
	wire.Bind(new(commentService.EventPublisher), new(*eventpublisher.Publisher)),
	repository.FileSet,
	handler.WebSet,

	repository.UserSet,
	repository.AuthSet,
	service.UserSet,
	service.AuthSet,
	handler.UserSet,
	handler.AuthSet,

	repository.EchoSet,
	service.EchoSet,
	handler.EchoSet,
	repository.CommentSet,
	service.CommentSet,
	handler.CommentSet,

	repository.CommonSet,
	service.FileSet,
	handler.FileSet,
	repository.InitSet,
	service.InitSet,
	handler.InitSet,
	service.CommonSet,
	handler.CommonSet,

	repository.WebhookSet,
	repository.KeyValueSet,

	repository.SettingSet,
	service.SettingSet,
	handler.SettingSet,

	repository.ConnectSet,
	service.ConnectSet,
	handler.ConnectSet,

	service.DashboardSet,
	handler.DashboardSet,

	repository.EmbeddingSet,
	service.EmbeddingSet,
	handler.EmbeddingSet,

	service.CopilotSet,
	// Copilot 的 UserReader 跨域绑定到 user 服务（取当前对话用户：展示名 + 检索按作者收口）。
	wire.Bind(new(copilotService.UserReader), new(*userService.UserService)),
	handler.CopilotSet,

	service.BackupSet,
	handler.BackupSet,
	service.MigratorSet,
	handler.MigrationSet,

	handler.MCPSet,

	handler.NewBundle,
)

var MiddlewareSet = wire.NewSet(
	repository.AuthSet,
	middleware.ProviderSet,
)

var TaskerSet = wire.NewSet(
	eventpublisher.New,
	repository.FileSet,
	repository.KeyValueSet,
	repository.WebhookSet,

	// SettingService 需要 TokenRevoker (管理员删 token 时写黑名单)，
	// 而 TokenRevoker 由 AuthSet 提供，因此 Tasker 也得包含一份。
	repository.AuthSet,
	repository.SettingSet,
	service.SettingSet,

	repository.EchoSet,
	service.EchoSet,

	repository.CommonSet,
	service.FileSet,
	service.CommonSet,

	repository.QueueSet,
	repository.VisitorSet,
	scheduled.ProviderSet,
	ProvideTaskManager,
)

var MigratorSet = wire.NewSet(
	migrator.ProviderSet,
)

// BuildApp 构建 Web 生命周期应用。
func BuildApp() (*app.App, error) {
	wire.Build(
		InfraSet,
		VisitorSet,
		StorageSet,
		DomainSet,
		RuntimeSet,
		AppSet,
	)
	return &app.App{}, nil
}

func BuildEventRegistrar(
	dbProvider func() *gorm.DB,
	ebProvider func() *busen.Bus,
	appCache cache.ICache[string, any],
	tx transaction.Transactor,
	backupScheduleApplier eventsubscriber.BackupScheduleApplier,
) (*eventregistry.EventRegistrar, error) {
	wire.Build(EventSet)
	return &eventregistry.EventRegistrar{}, nil
}

// BuildHandlers 使用 wire 生成的代码来构建 Handlers 实例。
// tracker 由顶层 BuildApp/BuildServer 注入,保证整个进程只有一个 visitor.Tracker 实例。
func BuildHandlers(
	dbProvider func() *gorm.DB,
	appCache cache.ICache[string, any],
	tx transaction.Transactor,
	ebProvider func() *busen.Bus,
	tracker *visitor.Tracker,
	jobManager *job.Manager,
	storageManager *storage.Manager,
) (*handler.Bundle, error) {
	wire.Build(HandlerSet)
	return &handler.Bundle{}, nil
}

// BuildJobManager 装配共享单例 *job.Manager：repo + 各领域 Runner（含其依赖的领域
// service），在构造期注册完成。Runner 依赖的 EmbeddingService / migrator.Importer 均不
// 含 *job.Manager，故无构造环。storageManager 由顶层共享单例注入，确保迁移导入 S3
// 设置时 reload 的就是文件服务在用的那份 Manager。
func BuildJobManager(
	dbProvider func() *gorm.DB,
	appCache cache.ICache[string, any],
	storageManager *storage.Manager,
) (*job.Manager, error) {
	wire.Build(
		repository.JobSet,
		// ReindexRunner ← EmbeddingService
		repository.EmbeddingSet,
		repository.EchoSet,
		repository.KeyValueSet,
		service.EmbeddingSet,
		// MigrationRunner ← migrator.Importer（无状态导入，不含 *job.Manager）
		migratorService.NewImporter,
		jobRunner.ProviderSet,
		ProvideJobManager,
	)
	return nil, nil
}

// BuildMiddlewares 构建中间件依赖。
func BuildMiddlewares(
	dbProvider func() *gorm.DB,
	appCache cache.ICache[string, any],
) (*middleware.Deps, error) {
	wire.Build(MiddlewareSet)
	return &middleware.Deps{}, nil
}

// BuildServer 构建 HTTP server
func BuildServer() (*server.Server, error) {
	wire.Build(
		InfraSet,
		VisitorSet,
		StorageSet,
		BuildJobManager,
		BuildHandlers,
		BuildMiddlewares,
		server.ProviderSet,
	)
	return &server.Server{}, nil
}

func BuildTasker(
	dbProvider func() *gorm.DB,
	appCache cache.ICache[string, any],
	tx transaction.Transactor,
	ebProvider func() *busen.Bus,
	tracker *visitor.Tracker,
	storageManager *storage.Manager,
) (*task.Manager, error) {
	wire.Build(TaskerSet)
	return &task.Manager{}, nil
}

func BuildMigrator(
	dbProvider func() *gorm.DB,
	appCache cache.ICache[string, any],
	tx transaction.Transactor,
) (*migrator.Worker, error) {
	wire.Build(MigratorSet)
	return &migrator.Worker{}, nil
}

// ProvideBackupScheduleApplier 从 task.Manager 中按能力取出实现了 BackupScheduleApplier
// 的那个 Task（即 *scheduled.Backup），供 BackupScheduler 订阅者在运行期重配备份计划。
// 取的是 Manager 持有的同一实例，故 Schedule 时捕获的 scheduler 对 Apply 可见。
func ProvideBackupScheduleApplier(m *task.Manager) eventsubscriber.BackupScheduleApplier {
	applier, ok := task.Find[eventsubscriber.BackupScheduleApplier](m)
	if !ok {
		panic("no scheduled task implements BackupScheduleApplier")
	}
	return applier
}

func ProvideSubscriptionProviders(
	dlr *eventsubscriber.DeadLetterResolver,
	bs *eventsubscriber.BackupScheduler,
	ap *eventsubscriber.AgentProcessor,
	ep *eventsubscriber.EmbeddingProcessor,
) []eventregistry.SubscriptionProvider {
	return []eventregistry.SubscriptionProvider{dlr, bs, ap, ep}
}
