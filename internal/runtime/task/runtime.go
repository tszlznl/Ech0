package task

import (
	"context"

	"github.com/lin-snow/ech0/internal/task"
)

// Runtime 适配 Tasker 到应用生命周期接口。
type Runtime struct {
	tasker *task.Tasker
}

func New(tasker *task.Tasker) *Runtime {
	return &Runtime{tasker: tasker}
}

func (r *Runtime) Name() string {
	return "task"
}

func (r *Runtime) Start(context.Context) error {
	return r.tasker.Start()
}

func (r *Runtime) Stop(context.Context) error {
	return r.tasker.Stop()
}
