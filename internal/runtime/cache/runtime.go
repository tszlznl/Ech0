package cache

import "context"

// ShutdownHook 用于在停止阶段执行缓存清理。
type ShutdownHook struct {
	cleanup func() error
}

func New(cleanup func() error) *ShutdownHook {
	return &ShutdownHook{cleanup: cleanup}
}

func (h *ShutdownHook) Name() string {
	return "cache_cleanup"
}

func (h *ShutdownHook) Shutdown(context.Context) error {
	if h.cleanup == nil {
		return nil
	}
	return h.cleanup()
}
