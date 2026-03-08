package app

import (
	"github.com/google/wire"
	bus "github.com/lin-snow/ech0/internal/event/bus"
	registry "github.com/lin-snow/ech0/internal/event/registry"
	"github.com/lin-snow/ech0/internal/server"
	"github.com/lin-snow/ech0/internal/task"
)

func ProvideComponents(
	_ *registry.RegisteredRegistrar,
	eventBus *bus.Component,
	tasker *task.Tasker,
	httpServer *server.Server,
) []Component {
	// 启动顺序：event bus -> task -> http；停止时 reverse，确保 bus 最后关闭。
	return []Component{eventBus, tasker, httpServer}
}

var ProviderSet = wire.NewSet(ProvideComponents, bus.NewComponent, NewApp)
