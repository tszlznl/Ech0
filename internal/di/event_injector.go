//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	"github.com/lin-snow/ech0/internal/cache"
	"github.com/lin-snow/ech0/internal/event"
	"github.com/lin-snow/ech0/internal/transaction"
	"gorm.io/gorm"
)

func BuildEventRegistrar(
	dbProvider func() *gorm.DB,
	ebProvider func() event.IEventBus,
	cacheFactory *cache.CacheFactory,
	tmFactory *transaction.TransactionManagerFactory,
) (*event.EventRegistrar, error) {
	wire.Build(
		EchoSet,
		UserSet,
		TodoSet,
		InboxSet,
		CacheSet,
		TransactionManagerSet,
		KeyValueSet,
		QueueSet,
		WebhookSet,
		EventSet,
	)
	return &event.EventRegistrar{}, nil
}
