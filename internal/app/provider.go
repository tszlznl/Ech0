package app

import (
	"github.com/google/wire"
	bus "github.com/lin-snow/ech0/internal/event/bus"
	registry "github.com/lin-snow/ech0/internal/event/registry"
	"github.com/lin-snow/ech0/internal/server"
	"github.com/lin-snow/ech0/internal/task"
)

func ProvideComponents(
	// 初始化哨兵依赖：强制触发事件订阅注册与对应 cleanup 挂载。
	_ *registry.RegisteredRegistrar,
	eventBus *bus.EventBus,
	tasker *task.Tasker,
	httpServer *server.Server,
) []Component {
	// 启动顺序：event bus -> task -> http；停止时 reverse，确保 bus 最后关闭。
	return []Component{eventBus, tasker, httpServer}
}

var ProviderSet = wire.NewSet(ProvideComponents, bus.NewEventBus, NewApp)
