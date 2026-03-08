package event

import (
	"sync"

	"github.com/google/wire"
	busen "github.com/lin-snow/Busen"
)

type RegisteredRegistrar struct {
	Registrar *EventRegistrar
}

func ProvideEventBusProvider() func() *busen.Bus {
	var once sync.Once
	var bus *busen.Bus
	return func() *busen.Bus {
		once.Do(func() {
			bus = NewBus()
		})
		return bus
	}
}

func ProvideRegisteredRegistrar(registrar *EventRegistrar) (*RegisteredRegistrar, func(), error) {
	if err := registrar.Register(); err != nil {
		return nil, nil, err
	}
	return &RegisteredRegistrar{Registrar: registrar}, func() {
		_ = registrar.Stop()
	}, nil
}

var ProviderSet = wire.NewSet(ProvideEventBusProvider, NewPublisher)
