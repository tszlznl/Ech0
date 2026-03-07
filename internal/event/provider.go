package event

import (
	"sync"

	"github.com/google/wire"
)

func ProvideEventBusProvider() func() IEventBus {
	var once sync.Once
	return func() IEventBus {
		once.Do(InitEventBus)
		return GetEventBus()
	}
}

var ProviderSet = wire.NewSet(ProvideEventBusProvider)
