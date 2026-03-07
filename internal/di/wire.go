//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	"github.com/lin-snow/ech0/internal/app"
	"github.com/lin-snow/ech0/internal/cache"
	"github.com/lin-snow/ech0/internal/event"
	"github.com/lin-snow/ech0/internal/handler"
	runtimeCache "github.com/lin-snow/ech0/internal/runtime/cache"
	runtimeEvent "github.com/lin-snow/ech0/internal/runtime/event"
	runtimeHTTP "github.com/lin-snow/ech0/internal/runtime/http"
	runtimeTask "github.com/lin-snow/ech0/internal/runtime/task"
	"github.com/lin-snow/ech0/internal/task"
	"github.com/lin-snow/ech0/internal/transaction"
	"gorm.io/gorm"
)

var AppSet = wire.NewSet(
	ProvideComponents,
	ProvideShutdownHooks,
	app.ProviderSet,
)

var DomainSet = wire.NewSet(
	ProvideHandlers,
	ProvideTasker,
	ProvideEventRegistrar,
)

var InfraSet = wire.NewSet(
	ProvideDBProvider,
	ProvideEventBusProvider,
	cache.ProviderSet,
	transaction.ProviderSet,
	ProvideGinEngine,
)

var RuntimeSet = wire.NewSet(
	ProvideHTTPServer,
	runtimeHTTP.ProviderSet,
	runtimeEvent.ProviderSet,
	runtimeTask.ProviderSet,
	runtimeCache.ProviderSet,
)

// BuildWebApp 构建 Web 生命周期应用。
func BuildWebApp() (*app.App, func(), error) {
	wire.Build(
		InfraSet,
		DomainSet,
		RuntimeSet,
		AppSet,
	)
	return &app.App{}, nil, nil
}

// BuildApp 兼容旧入口，委托给 BuildWebApp。
func BuildApp() (*app.App, func(), error) {
	return BuildWebApp()
}

func BuildEventRegistrar(
	dbProvider func() *gorm.DB,
	ebProvider func() event.IEventBus,
	appCache cache.ICache[string, any],
	tx transaction.Transactor,
) (*event.EventRegistrar, error) {
	wire.Build(
		EventGraphSet,
	)
	return &event.EventRegistrar{}, nil
}

// BuildHandlers 使用 wire 生成的代码来构建 Handlers 实例。
func BuildHandlers(
	dbProvider func() *gorm.DB,
	appCache cache.ICache[string, any],
	tx transaction.Transactor,
	ebProvider func() event.IEventBus,
) (*handler.Bundle, error) {
	wire.Build(
		HandlerGraphSet,
	)
	return &handler.Bundle{}, nil
}

// BuildWebRuntime 构建 HTTP runtime（用于测试和独立启动场景）。
func BuildWebRuntime() (*runtimeHTTP.Runtime, error) {
	wire.Build(
		InfraSet,
		DomainSet,
		ProvideHTTPServer,
		runtimeHTTP.ProviderSet,
	)
	return &runtimeHTTP.Runtime{}, nil
}

func BuildTasker(
	dbProvider func() *gorm.DB,
	appCache cache.ICache[string, any],
	tx transaction.Transactor,
	ebProvider func() event.IEventBus,
) (*task.Tasker, error) {
	wire.Build(
		TaskerGraphSet,
	)
	return &task.Tasker{}, nil
}
