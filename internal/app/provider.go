package app

import (
	"github.com/google/wire"
	runtimeCache "github.com/lin-snow/ech0/internal/runtime/cache"
	runtimeEvent "github.com/lin-snow/ech0/internal/runtime/event"
	runtimeHTTP "github.com/lin-snow/ech0/internal/runtime/http"
	runtimeTask "github.com/lin-snow/ech0/internal/runtime/task"
)

func ProvideLifecycles(
	eventRuntime *runtimeEvent.Runtime,
	taskRuntime *runtimeTask.Runtime,
	httpRuntime *runtimeHTTP.Runtime,
	cacheRuntime *runtimeCache.Runtime,
) []Lifecycle {
	// 启动顺序必须保持为 event -> task -> http：
	// 1. 先注册事件处理器，避免后台任务启动后发布的事件没有订阅者。
	// 2. 再启动任务调度器，让任务依赖的事件流已经可用。
	// 3. 再开放 HTTP 入口，避免服务对外可见时后台基础设施尚未完成初始化。
	// cache cleanup 仅在 Stop 阶段执行，其 Start 为 no-op，因此放在尾部即可在停止时最先清理。
	return []Lifecycle{eventRuntime, taskRuntime, httpRuntime, cacheRuntime}
}

var ProviderSet = wire.NewSet(ProvideLifecycles, NewApp)
