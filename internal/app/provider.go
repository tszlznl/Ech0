// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package app

import (
	"context"

	"github.com/google/wire"
	bus "github.com/lin-snow/ech0/internal/event/bus"
	"github.com/lin-snow/ech0/internal/job"
	"github.com/lin-snow/ech0/internal/kvstore"
	"github.com/lin-snow/ech0/internal/server"
	"github.com/lin-snow/ech0/internal/setting"
	"github.com/lin-snow/ech0/internal/task"
)

func ProvideOptions(
	registrar *bus.EventRegistrar,
	jobManager *job.Manager,
	taskManager *task.Manager,
	httpServer *server.Server,
	durableKV kvstore.Store,
) []Option {
	return []Option{
		// jobManager 排在 httpServer 前：其 Start 做启动期孤儿清理，须先于对外服务。
		// Runner 已在构造期装配进 jobManager、Task 已装配进 taskManager，无需额外注册步骤。
		Components(jobManager, taskManager, httpServer),
		// 启动期一次性副作用，按序执行且必早于任何 component.Start：
		// 先 seed 缺失的配置 key（此后各读路径直接命中，Get 不再承担「读时 seed」副作用），
		// 再注册事件订阅。
		BeforeStart(func(ctx context.Context) error {
			if err := setting.Seed(ctx, durableKV); err != nil {
				return err
			}
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

var ProviderSet = wire.NewSet(ProvideOptions, NewApp)
