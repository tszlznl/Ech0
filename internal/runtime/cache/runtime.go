package cache

import "context"

// Runtime 用于在停止阶段执行缓存清理。
type Runtime struct {
	cleanup func() error
}

func New(cleanup func() error) *Runtime {
	return &Runtime{cleanup: cleanup}
}

func (r *Runtime) Name() string {
	return "cache_cleanup"
}

func (r *Runtime) Start(context.Context) error {
	return nil
}

func (r *Runtime) Stop(context.Context) error {
	if r.cleanup == nil {
		return nil
	}
	return r.cleanup()
}

func (r *Runtime) Healthy(context.Context) error {
	return nil
}
