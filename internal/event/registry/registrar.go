package registry

import (
	"context"
	"sync/atomic"

	busen "github.com/lin-snow/Busen"
	contracts "github.com/lin-snow/ech0/internal/event/contracts"
)

type WebhookObserver interface {
	HandleObservation(ctx context.Context, obs contracts.WebhookObservation) error
	Stop()
	Wait()
}

type EventRegistrar struct {
	bus        *busen.Bus
	observer   WebhookObserver
	providers  []SubscriptionProvider
	unsub      []func()
	registered atomic.Bool
}

func NewEventRegistry(
	busProvider func() *busen.Bus,
	observer WebhookObserver,
	providers []SubscriptionProvider,
) *EventRegistrar {
	return &EventRegistrar{
		bus:       busProvider(),
		observer:  observer,
		providers: providers,
	}
}

func (er *EventRegistrar) Register() error {
	if er.registered.Load() {
		return nil
	}

	for _, provider := range er.providers {
		if provider == nil {
			continue
		}
		for _, subscription := range provider.Subscriptions() {
			unsub, err := subscription.Register(er.bus)
			if err != nil {
				er.stopSubscriptions()
				return err
			}
			er.unsub = append(er.unsub, unsub)
		}
	}

	for _, subscription := range webhookSubscriptions(er.observer) {
		unsub, err := subscription.Register(er.bus)
		if err != nil {
			er.stopSubscriptions()
			return err
		}
		er.unsub = append(er.unsub, unsub)
	}

	er.registered.Store(true)
	return nil
}

func (er *EventRegistrar) Stop() error {
	if !er.registered.Load() {
		return nil
	}
	er.stopSubscriptions()
	er.observer.Stop()
	er.observer.Wait()
	return nil
}

func (er *EventRegistrar) stopSubscriptions() {
	for i := len(er.unsub) - 1; i >= 0; i-- {
		er.unsub[i]()
	}
	er.unsub = nil
}
