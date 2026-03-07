package app

import "context"

// Component 定义应用生命周期组件。
type Component interface {
	Name() string
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// ShutdownHook 定义仅在应用退出时执行的资源清理钩子。
type ShutdownHook interface {
	Name() string
	Shutdown(ctx context.Context) error
}
