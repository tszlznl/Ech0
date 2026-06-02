// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package app

import (
	"context"

	"github.com/google/wire"
	bus "github.com/lin-snow/ech0/internal/event/bus"
	registry "github.com/lin-snow/ech0/internal/event/registry"
	"github.com/lin-snow/ech0/internal/job"
	"github.com/lin-snow/ech0/internal/migrator"
	"github.com/lin-snow/ech0/internal/server"
	"github.com/lin-snow/ech0/internal/task"
)

func ProvideOptions(
	registrar *registry.EventRegistrar,
	eventBus *bus.EventBus,
	jobManager *job.Manager,
	taskManager *task.Manager,
	migratorWorker *migrator.Worker,
	httpServer *server.Server,
) []Option {
	return []Option{
		// jobManager 排在 httpServer 前：其 Start 做启动期孤儿清理，须先于对外服务。
		// Runner 已在构造期装配进 jobManager、Task 已装配进 taskManager，无需额外注册步骤。
		Components(eventBus, jobManager, taskManager, migratorWorker, httpServer),
		BeforeStart(func(context.Context) error {
			return registrar.Register()
		}),
		AfterStop(func(context.Context) error {
			return registrar.Stop()
		}),
	}
}

func NewApp(opts []Option) *App {
	return New(opts...)
}

var ProviderSet = wire.NewSet(ProvideOptions, bus.NewEventBus, NewApp)
