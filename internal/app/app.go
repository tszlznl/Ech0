package app

import (
	"context"
	"errors"
	"reflect"
	"sync"
)

// App 是应用组件编排器。
type App struct {
	mu sync.Mutex

	components []Component

	running bool
}

// NewApp 创建应用组件编排器。
func NewApp(components []Component) *App {
	return &App{
		components: components,
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
	for _, component := range a.components {
		if component == nil {
			return &AppError{
				Code:      CodeDependencyMissing,
				Op:        "app.start",
				Component: "nil_component",
			}
		}
		if err := component.Start(ctx); err != nil {
			a.stopComponentsReverse(ctx, started)
			return &AppError{
				Code:      CodeComponentStartFailed,
				Op:        "app.start",
				Component: componentName(component),
				Cause:     err,
			}
		}
		started = append(started, component)
	}

	a.running = true
	return nil
}

// Stop 按反向顺序停止应用组件单元。
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
		component := a.components[i]
		if component == nil {
			continue
		}
		if err := component.Stop(ctx); err != nil {
			errs = append(errs, &AppError{
				Code:      CodeComponentStopFailed,
				Op:        "app.stop",
				Component: componentName(component),
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

func componentName(component Component) string {
	if component == nil {
		return "nil_component"
	}
	if namer, ok := component.(Namer); ok {
		return namer.Name()
	}

	t := reflect.TypeOf(component)
	if t == nil {
		return "unknown_component"
	}
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Name() != "" {
		return t.Name()
	}
	return "unknown_component"
}
