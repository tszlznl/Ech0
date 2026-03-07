package cache

import "context"

// Runtime 适配缓存清理到统一生命周期接口。
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
