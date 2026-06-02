// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package job 是长时有状态作业的通用框架：统一承载状态机、goroutine 生命周期、
// 取消、持久化与状态查询；各领域只实现 Runner，触发器只调 Submit。
//
// 本包仅依赖 internal/model/job 与标准库，不 import 任何领域 service；领域 Runner
// 放在子包 internal/job/runner，故无 import 环。
package job

import (
	"context"
	"errors"

	jobModel "github.com/lin-snow/ech0/internal/model/job"
)

// ErrNotFound 表示该类型当前没有作业行。上层据此合成领域哨兵（如 migration 的 idle）。
var ErrNotFound = errors.New("job not found")

// ReportFunc 供 Runner 上报实时进度（仅进内存，不落库）。phase 必填；snapshot 可为
// nil，表示只更新阶段、不覆盖 durable Payload。
type ReportFunc func(phase string, snapshot any)

// Runner 是作业的工作单元：payload 为 Submit 传入的原始 JSON；返回的 result 作为终态
// Payload 落库（nil 则保留原 payload）；返回 error 置 failed；必须在长循环里检查
// ctx 取消。作者端不直接实现它，而用 Adapt 适配 typed 的工作函数。
type Runner interface {
	Run(ctx context.Context, payload []byte, report ReportFunc) (result any, err error)
}

// JobRepository 是 jobs 表的持久化抽象（每类型单行）。
type JobRepository interface {
	Upsert(ctx context.Context, j *jobModel.Job) error
	// GetByType 查无返回 (零值, ErrNotFound)。
	GetByType(ctx context.Context, jobType string) (jobModel.Job, error)
	// SweepRunning 把残留的 pending/running 行批量置 failed（启动期孤儿清理）。
	SweepRunning(ctx context.Context, reason string) error
	Delete(ctx context.Context, jobType string) error
}

// Progress 是内存态的实时进度。
type Progress struct {
	Phase    string
	Snapshot any
}
