// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package registry

import (
	"context"

	busen "github.com/lin-snow/Busen"
)

type SubscriptionProvider interface {
	Subscriptions() []Subscription
}

type Subscription struct {
	register func(*busen.Bus) (func(), error)
}

func (s Subscription) Register(bus *busen.Bus) (func(), error) {
	return s.register(bus)
}

func TypedSubscription[T any](
	handler func(context.Context, T) error,
	opts ...busen.SubscribeOption,
) Subscription {
	return Subscription{
		register: func(bus *busen.Bus) (func(), error) {
			return busen.Subscribe(bus,
				func(ctx context.Context, e busen.Event[T]) error {
					return handler(ctx, e.Value)
				},
				opts...,
			)
		},
	}
}

func TopicSubscription[T any](
	pattern string,
	handler func(context.Context, T) error,
	opts ...busen.SubscribeOption,
) Subscription {
	return Subscription{
		register: func(bus *busen.Bus) (func(), error) {
			return busen.SubscribeTopic(bus, pattern,
				func(ctx context.Context, e busen.Event[T]) error {
					return handler(ctx, e.Value)
				},
				opts...,
			)
		},
	}
}
