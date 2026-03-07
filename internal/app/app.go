package app

import (
	"context"
	"errors"
	"sync"
)

// App 是应用生命周期编排器。
type App struct {
	mu sync.Mutex

	components    []Component
	shutdownHooks []ShutdownHook

	running bool
}

// NewApp 创建应用生命周期编排器。
func NewApp(components []Component, shutdownHooks []ShutdownHook) *App {
	return &App{
		components:    components,
		shutdownHooks: shutdownHooks,
	}
}

// Start 按顺序启动应用组件链。
func (a *App) Start(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.running {
		return &AppError{
			Code:      CodeInvalidState,
			Op:        "app.start",
			Component: "app",
		}
	}
	if len(a.components) == 0 {
		return &AppError{
			Code:      CodeDependencyMissing,
			Op:        "app.start",
			Component: "components",
		}
	}

	started := make([]Component, 0, len(a.components))
	for _, c := range a.components {
		if c == nil {
			return &AppError{
				Code:      CodeDependencyMissing,
				Op:        "app.start",
				Component: "nil_component",
			}
		}
		if err := c.Start(ctx); err != nil {
			a.stopComponentsReverse(ctx, started)
			return &AppError{
				Code:      CodeComponentStartFailed,
				Op:        "app.start",
				Component: c.Name(),
				Cause:     err,
			}
		}
		started = append(started, c)
	}

	a.running = true
	return nil
}

// Stop 按反向顺序停止应用组件链，并在最后执行退出钩子。
func (a *App) Stop(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.running {
		return &AppError{
			Code:      CodeInvalidState,
			Op:        "app.stop",
			Component: "app",
		}
	}

	var errs []error
	for i := len(a.components) - 1; i >= 0; i-- {
		c := a.components[i]
		if c == nil {
			continue
		}
		if err := c.Stop(ctx); err != nil {
			errs = append(errs, &AppError{
				Code:      CodeComponentStopFailed,
				Op:        "app.stop",
				Component: c.Name(),
				Cause:     err,
			})
		}
	}

	for _, hook := range a.shutdownHooks {
		if hook == nil {
			continue
		}
		if err := hook.Shutdown(ctx); err != nil {
			errs = append(errs, &AppError{
				Code:      CodeComponentStopFailed,
				Op:        "app.stop",
				Component: hook.Name(),
				Cause:     err,
			})
		}
	}

	a.running = false
	return errors.Join(errs...)
}

// StopAll 停止所有已运行组件。
func (a *App) StopAll(ctx context.Context) error {
	var errs []error

	if a.IsRunning() {
		if err := a.Stop(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// IsRunning 返回应用组件链是否运行中。
func (a *App) IsRunning() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.running
}

func (a *App) stopComponentsReverse(ctx context.Context, components []Component) {
	for i := len(components) - 1; i >= 0; i-- {
		_ = components[i].Stop(ctx)
	}
}
