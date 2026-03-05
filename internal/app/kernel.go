package app

import (
	"context"
	"errors"
	"sync"
)

// Kernel 是应用生命周期编排器。
type Kernel struct {
	mu sync.Mutex

	webComponents []Component
	sshComponent  Component

	webRunning bool
	sshRunning bool
}

// NewKernel 创建应用内核。
func NewKernel(webComponents []Component, sshComponent Component) *Kernel {
	return &Kernel{
		webComponents: webComponents,
		sshComponent:  sshComponent,
	}
}

// StartWeb 按顺序启动 Web 组件链。
func (k *Kernel) StartWeb(ctx context.Context) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.webRunning {
		return &AppError{
			Code:      CodeInvalidState,
			Op:        "kernel.start_web",
			Component: "web",
		}
	}
	if len(k.webComponents) == 0 {
		return &AppError{
			Code:      CodeDependencyMissing,
			Op:        "kernel.start_web",
			Component: "web_components",
		}
	}

	started := make([]Component, 0, len(k.webComponents))
	for _, c := range k.webComponents {
		if c == nil {
			return &AppError{
				Code:      CodeDependencyMissing,
				Op:        "kernel.start_web",
				Component: "nil_component",
			}
		}
		if err := c.Start(ctx); err != nil {
			k.stopReverse(ctx, started)
			return &AppError{
				Code:      CodeComponentStartFailed,
				Op:        "kernel.start_web",
				Component: c.Name(),
				Cause:     err,
			}
		}
		started = append(started, c)
	}

	k.webRunning = true
	return nil
}

// StopWeb 按反向顺序停止 Web 组件链。
func (k *Kernel) StopWeb(ctx context.Context) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if !k.webRunning {
		return &AppError{
			Code:      CodeInvalidState,
			Op:        "kernel.stop_web",
			Component: "web",
		}
	}

	var errs []error
	for i := len(k.webComponents) - 1; i >= 0; i-- {
		c := k.webComponents[i]
		if c == nil {
			continue
		}
		if err := c.Stop(ctx); err != nil {
			errs = append(errs, &AppError{
				Code:      CodeComponentStopFailed,
				Op:        "kernel.stop_web",
				Component: c.Name(),
				Cause:     err,
			})
		}
	}

	k.webRunning = false
	return errors.Join(errs...)
}

// StartSSH 启动 SSH 组件。
func (k *Kernel) StartSSH(ctx context.Context) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.sshComponent == nil {
		return &AppError{
			Code:      CodeDependencyMissing,
			Op:        "kernel.start_ssh",
			Component: "ssh",
		}
	}
	if k.sshRunning {
		return &AppError{
			Code:      CodeInvalidState,
			Op:        "kernel.start_ssh",
			Component: "ssh",
		}
	}
	if err := k.sshComponent.Start(ctx); err != nil {
		return &AppError{
			Code:      CodeComponentStartFailed,
			Op:        "kernel.start_ssh",
			Component: k.sshComponent.Name(),
			Cause:     err,
		}
	}

	k.sshRunning = true
	return nil
}

// StopSSH 停止 SSH 组件。
func (k *Kernel) StopSSH(ctx context.Context) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.sshComponent == nil {
		return &AppError{
			Code:      CodeDependencyMissing,
			Op:        "kernel.stop_ssh",
			Component: "ssh",
		}
	}
	if !k.sshRunning {
		return &AppError{
			Code:      CodeInvalidState,
			Op:        "kernel.stop_ssh",
			Component: "ssh",
		}
	}
	if err := k.sshComponent.Stop(ctx); err != nil {
		return &AppError{
			Code:      CodeComponentStopFailed,
			Op:        "kernel.stop_ssh",
			Component: k.sshComponent.Name(),
			Cause:     err,
		}
	}

	k.sshRunning = false
	return nil
}

// StopAll 停止所有已运行组件。
func (k *Kernel) StopAll(ctx context.Context) error {
	var errs []error

	if k.IsSSHRunning() {
		if err := k.StopSSH(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	if k.IsWebRunning() {
		if err := k.StopWeb(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// IsWebRunning 返回 Web 组件链是否运行中。
func (k *Kernel) IsWebRunning() bool {
	k.mu.Lock()
	defer k.mu.Unlock()
	return k.webRunning
}

// IsSSHRunning 返回 SSH 组件是否运行中。
func (k *Kernel) IsSSHRunning() bool {
	k.mu.Lock()
	defer k.mu.Unlock()
	return k.sshRunning
}

func (k *Kernel) stopReverse(ctx context.Context, components []Component) {
	for i := len(components) - 1; i >= 0; i-- {
		_ = components[i].Stop(ctx)
	}
}
