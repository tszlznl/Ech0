package app

import (
	"context"

	"github.com/google/wire"
	bus "github.com/lin-snow/ech0/internal/event/bus"
	registry "github.com/lin-snow/ech0/internal/event/registry"
	"github.com/lin-snow/ech0/internal/migrator"
	"github.com/lin-snow/ech0/internal/server"
	"github.com/lin-snow/ech0/internal/task"
)

func ProvideOptions(
	registrar *registry.EventRegistrar,
	eventBus *bus.EventBus,
	tasker *task.Tasker,
	migratorWorker *migrator.Worker,
	httpServer *server.Server,
) []Option {
	return []Option{
		Components(eventBus, tasker, migratorWorker, httpServer),
		BeforeStart(func(context.Context) error {
			return registrar.Register()
		}),
		AfterStop(func(context.Context) error {
			return registrar.Stop()
		}),
	}
}

func NewApp(opts []Option) *App {
	return New(opts...)
}

var ProviderSet = wire.NewSet(ProvideOptions, bus.NewEventBus, NewApp)
