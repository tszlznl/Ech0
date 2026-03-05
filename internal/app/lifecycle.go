package app

import "context"

// Component 定义应用生命周期组件。
type Component interface {
	Name() string
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Healthy(ctx context.Context) error
}
