// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package bus

import (
	"context"
	"sync/atomic"

	"github.com/lin-snow/ech0/internal/event"
	"github.com/lin-snow/ech0/pkg/busen"
)

// WebhookObserver 接收已发生事件的中立观察并分发给 webhook（由 internal/webhook.Dispatcher 实现）。
type WebhookObserver interface {
	HandleObservation(ctx context.Context, obs event.WebhookObservation) error
	Stop()
	Wait()
}

// EventRegistrar 在启动时把所有领域订阅者与 webhook 桥接注册到总线，并在停机时统一拆除。
type EventRegistrar struct {
	bus         *busen.Bus
	observer    WebhookObserver
	subscribers []Subscriber
	unsub       []func()
	registered  atomic.Bool
}

func NewEventRegistry(
	busProvider func() *busen.Bus,
	observer WebhookObserver,
	subscribers []Subscriber,
) *EventRegistrar {
	return &EventRegistrar{
		bus:         busProvider(),
		observer:    observer,
		subscribers: subscribers,
	}
}

func (er *EventRegistrar) Register() error {
	if er.registered.Load() {
		return nil
	}

	for _, sub := range er.subscribers {
		if sub == nil {
			continue
		}
		for _, reg := range sub.Registrations() {
			unsub, err := reg(er.bus)
			if err != nil {
				er.stopSubscriptions()
				return err
			}
			er.unsub = append(er.unsub, unsub)
		}
	}

	webhookUnsubs, err := registerWebhookObservers(er.bus, er.observer.HandleObservation)
	if err != nil {
		er.stopSubscriptions()
		return err
	}
	er.unsub = append(er.unsub, webhookUnsubs...)

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
