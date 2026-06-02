// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package job

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	jobModel "github.com/lin-snow/ech0/internal/model/job"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"go.uber.org/zap"
)

const logModule = "job"

var (
	// ErrNoRunner 提交了未注册类型的作业。
	ErrNoRunner = errors.New("no runner registered for job type")
	// ErrAlreadyRunning 该类型已有一条非终态作业（同类型互斥）。
	ErrAlreadyRunning = errors.New("a job of this type is already running")
)

// Manager 管理所有作业的生命周期：Runner 注册表、durable 持久化、内存实时进度、
// 取消句柄。它从不解析领域 payload，只搬运 JSON。实现 app.Component。
type Manager struct {
	repo JobRepository

	mu      sync.Mutex
	runners map[string]Runner
	live    map[string]*Progress
	cancels map[string]context.CancelFunc
}

func NewManager(repo JobRepository) *Manager {
	return &Manager{
		repo:    repo,
		runners: make(map[string]Runner),
		live:    make(map[string]*Progress),
		cancels: make(map[string]context.CancelFunc),
	}
}

// Register 登记某类型的 Runner，须在任何 Submit 之前于启动期调用。
func (m *Manager) Register(jobType string, r Runner) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.runners[jobType] = r
}

// Submit 提交一次作业：校验已注册、同类型互斥、upsert pending、登记取消句柄，
// 起 goroutine 执行，返回 pending 行。
func (m *Manager) Submit(ctx context.Context, jobType string, payload []byte) (jobModel.Job, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	runner, ok := m.runners[jobType]
	if !ok {
		return jobModel.Job{}, fmt.Errorf("%w: %s", ErrNoRunner, jobType)
	}

	// 同类型互斥：现存非终态行则拒绝。持锁读判断+upsert，单进程下即原子。
	if existing, err := m.repo.GetByType(ctx, jobType); err == nil {
		if !existing.Status.IsTerminal() {
			return jobModel.Job{}, ErrAlreadyRunning
		}
	} else if !errors.Is(err, ErrNotFound) {
		return jobModel.Job{}, err
	}

	now := time.Now().UTC().Unix()
	pending := jobModel.Job{
		Type:      jobType,
		Status:    jobModel.StatusPending,
		Payload:   string(payload),
		StartedAt: &now,
	}
	if err := m.repo.Upsert(ctx, &pending); err != nil {
		return jobModel.Job{}, err
	}

	// 作业独立于触发它的 HTTP 请求，用 background 派生可取消 ctx；持锁登记取消句柄，
	// 消除「pending 已建、cancel 未登记」的窗口。
	runCtx, cancel := context.WithCancel(context.Background())
	m.cancels[jobType] = cancel
	delete(m.live, jobType)

	go m.run(runCtx, jobType, runner, pending)

	logUtil.GetLogger().Info("job submitted", zap.String("module", logModule), zap.String("type", jobType))
	return pending, nil
}

// run 在独立 goroutine 内推进作业：running → success/failed/cancelled。
func (m *Manager) run(runCtx context.Context, jobType string, runner Runner, base jobModel.Job) {
	// durable 写用 background ctx，避免取消后终态行写不进去。
	dbCtx := context.Background()
	report := func(phase string, snapshot any) { m.setLive(jobType, phase, snapshot) }

	base.Status = jobModel.StatusRunning
	if err := m.repo.Upsert(dbCtx, &base); err != nil {
		logUtil.GetLogger().Error("job mark running failed",
			zap.String("module", logModule), zap.String("type", jobType), zap.Error(err))
	}

	result, runErr := runner.Run(runCtx, []byte(base.Payload), report)

	now := time.Now().UTC().Unix()
	base.FinishedAt = &now
	base.Phase = m.takeLivePhase(jobType)

	switch {
	case errors.Is(runCtx.Err(), context.Canceled):
		base.Status = jobModel.StatusCancelled
		base.Error = ""
		logUtil.GetLogger().Warn("job cancelled", zap.String("module", logModule), zap.String("type", jobType))
	case runErr != nil:
		base.Status = jobModel.StatusFailed
		base.Error = runErr.Error()
		logUtil.GetLogger().Error("job failed",
			zap.String("module", logModule), zap.String("type", jobType), zap.Error(runErr))
	default:
		base.Status = jobModel.StatusSuccess
		base.Error = ""
		if result != nil {
			base.Payload = mustJSON(result)
		}
		logUtil.GetLogger().Info("job succeeded", zap.String("module", logModule), zap.String("type", jobType))
	}

	if err := m.repo.Upsert(dbCtx, &base); err != nil {
		logUtil.GetLogger().Error("job persist terminal failed",
			zap.String("module", logModule), zap.String("type", jobType),
			zap.String("status", string(base.Status)), zap.Error(err))
	}

	m.clear(jobType)
}

// Get 返回 durable 行；本进程正在跑时叠加内存实时进度。snapshot 为 nil 时只覆盖
// Phase，不动 durable Payload。
func (m *Manager) Get(ctx context.Context, jobType string) (jobModel.Job, error) {
	row, err := m.repo.GetByType(ctx, jobType)
	if err != nil {
		return row, err
	}
	m.mu.Lock()
	p := m.live[jobType]
	m.mu.Unlock()
	if p != nil {
		row.Phase = p.Phase
		if p.Snapshot != nil {
			row.Payload = mustJSON(p.Snapshot)
		}
	}
	return row, nil
}

// Delete 删除该类型的行并清空内存进度，使其回到「无作业」。仅应在终态时调用。
func (m *Manager) Delete(ctx context.Context, jobType string) error {
	if err := m.repo.Delete(ctx, jobType); err != nil {
		return err
	}
	m.clear(jobType)
	return nil
}

// Cancel 触发该类型在跑作业的 ctx 取消；无在跑/已终态则 no-op。
func (m *Manager) Cancel(jobType string) error {
	m.mu.Lock()
	cancel := m.cancels[jobType]
	m.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	return nil
}

func (m *Manager) setLive(jobType, phase string, snapshot any) {
	m.mu.Lock()
	m.live[jobType] = &Progress{Phase: phase, Snapshot: snapshot}
	m.mu.Unlock()
}

func (m *Manager) takeLivePhase(jobType string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if p := m.live[jobType]; p != nil {
		return p.Phase
	}
	return ""
}

func (m *Manager) clear(jobType string) {
	m.mu.Lock()
	delete(m.live, jobType)
	delete(m.cancels, jobType)
	m.mu.Unlock()
}

// Name 实现 app.Namer。
func (m *Manager) Name() string { return "job" }

// Start 把上次进程残留的 pending/running 行扫成 failed，避免前端永久转圈。幂等。
func (m *Manager) Start(ctx context.Context) error {
	if err := m.repo.SweepRunning(ctx, "interrupted by restart"); err != nil {
		logUtil.GetLogger().Error("sweep orphan jobs failed", zap.String("module", logModule), zap.Error(err))
		return err
	}
	return nil
}

// Stop 取消所有在跑作业，使其协作退出。
func (m *Manager) Stop(context.Context) error {
	m.mu.Lock()
	cancels := make([]context.CancelFunc, 0, len(m.cancels))
	for _, c := range m.cancels {
		cancels = append(cancels, c)
	}
	m.mu.Unlock()
	for _, c := range cancels {
		c()
	}
	return nil
}

func mustJSON(v any) string {
	if v == nil {
		return ""
	}
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}
