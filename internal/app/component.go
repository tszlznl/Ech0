package app

import "context"

// Component 定义应用可启动/停止的组件单元。
type Component interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// Namer 为可选能力，用于生成更友好的组件错误信息。
type Namer interface {
	Name() string
}
