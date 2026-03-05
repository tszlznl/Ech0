//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	"github.com/lin-snow/ech0/internal/cache"
	"github.com/lin-snow/ech0/internal/event"
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/transaction"
	"gorm.io/gorm"
)

// BuildHandlers 使用wire生成的代码来构建Handlers实例
func BuildHandlers(
	dbProvider func() *gorm.DB,
	cacheFactory *cache.CacheFactory,
	tmFactory *transaction.TransactionManagerFactory,
	ebProvider func() event.IEventBus,
) (*handler.Bundle, error) {
	wire.Build(
		CacheSet,
		FsSet,
		StorageSet,
		TransactionManagerSet,
		WebSet,
		UserSet,
		EchoSet,
		CommonSet,
		WebhookSet,
		KeyValueSet,
		SettingSet,
		InboxSet,
		TodoSet,
		ConnectSet,
		MetricSet,
		MonitorSet,
		DashboardSet,
		AgentSet,
		BackupSet,
		handler.NewBundle,
	)
	return &handler.Bundle{}, nil
}
