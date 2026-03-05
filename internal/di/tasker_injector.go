//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	"github.com/lin-snow/ech0/internal/cache"
	"github.com/lin-snow/ech0/internal/event"
	"github.com/lin-snow/ech0/internal/task"
	"github.com/lin-snow/ech0/internal/transaction"
	"gorm.io/gorm"
)

func BuildTasker(
	dbProvider func() *gorm.DB,
	cacheFactory *cache.CacheFactory,
	tmFactory *transaction.TransactionManagerFactory,
	ebProvider func() event.IEventBus,
) (*task.Tasker, error) {
	wire.Build(
		CacheSet,
		KeyValueSet,
		TransactionManagerSet,
		WebhookSet,
		SettingSet,
		EchoSet,
		CommonSet,
		QueueSet,
		TaskSet,
	)
	return &task.Tasker{}, nil
}
