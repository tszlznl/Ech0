// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package task 是周期性后台作业的通用框架：统一承载共享调度器与调度生命周期，
// 各领域只实现 Task 把自己挂上调度器。
//
// 对应 internal/job：job 维护 map[string]Runner（一次性、可取消、有状态机与持久化），
// task 维护有序 []Task（周期触发、fire-and-forget）。两者共享「瘦 Manager + 子包注册」
// 的形态，但语义不同，故各用各的接口。
//
// 本包只依赖 gocron 与标准库，不 import 任何领域 service；领域 Task 放在子包
// internal/task/scheduled，故无 import 环。
package task

import (
	"context"

	"github.com/go-co-op/gocron/v2"
)

// Task 是一个可调度的工作单元：在启动期把自己的 cron/interval 作业挂到共享 scheduler 上。
type Task interface {
	// Name 返回任务标识，用于日志与按能力查找。
	Name() string
	// Schedule 把本任务的作业注册到共享 scheduler。在 Manager.Start 时按列表顺序调用一次。
	Schedule(ctx context.Context, s gocron.Scheduler) error
}

// StopHook 是可选能力：实现它的 Task 会在 Manager 优雅停机、关闭 scheduler 之前被回调，
// 用于补一次落盘（如 visitor 快照），避免进程在下次 cron 触发前停止导致当天数据丢失。
type StopHook interface {
	OnStop(ctx context.Context)
}
