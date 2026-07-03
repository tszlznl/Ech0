// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package task

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/go-co-op/gocron/v2"
	logUtil "github.com/lin-snow/ech0/pkg/log"
)

const logModule = "task"

// Manager 持有一组 Task 与底层 gocron 调度器，统一负责调度生命周期。实现 app.Component。
// 它从不关心某个 Task 具体做什么，只负责「启动期逐个挂上、停机前给 StopHook 补刀、关闭调度器」。
type Manager struct {
	scheduler gocron.Scheduler
	tasks     []Task

	mu      sync.Mutex
	started bool
}

// NewManager 建一个 UTC 调度器（与 visitor.Tracker 内部日期 key 对齐，避免时区错配导致
// 每日最后一段数据永远无法被快照），并持有给定的 Task 列表。
func NewManager(tasks ...Task) (*Manager, error) {
	scheduler, err := gocron.NewScheduler(gocron.WithLocation(time.UTC))
	if err != nil {
		return nil, err
	}
	return &Manager{scheduler: scheduler, tasks: tasks}, nil
}

// Name 实现 app.Namer。
func (m *Manager) Name() string { return "task" }

// Start 依次让每个 Task 把自己挂上调度器，再启动调度器。幂等。
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.started {
		return nil
	}
	if m.scheduler == nil {
		return errors.New("scheduler is nil")
	}

	for _, t := range m.tasks {
		if err := t.Schedule(ctx, m.scheduler); err != nil {
			logUtil.GetLogger().Error("failed to schedule task",
				slog.String("module", logModule), slog.String("task", t.Name()), logUtil.Err(err))
			return err
		}
	}

	m.scheduler.Start()
	m.started = true
	return nil
}

// Stop 先给实现 StopHook 的 Task 一次补刀机会，再关闭调度器。
func (m *Manager) Stop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.started || m.scheduler == nil {
		return nil
	}

	for _, t := range m.tasks {
		if h, ok := t.(StopHook); ok {
			h.OnStop(ctx)
		}
	}

	if err := m.scheduler.Shutdown(); err != nil {
		logUtil.GetLogger().Error("failed to shutdown scheduler",
			slog.String("module", logModule), logUtil.Err(err))
		return err
	}
	m.started = false
	return nil
}
