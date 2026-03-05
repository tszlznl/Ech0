//go:build wireinject
// +build wireinject

package di

import "github.com/google/wire"

var InfraSet = wire.NewSet(
	ProvideDBProvider,
	ProvideEventBusProvider,
	ProvideCacheFactory,
	ProvideCache,
	ProvideCacheCleanup,
	ProvideTransactionManagerFactory,
	ProvideTransactionManager,
	ProvideGinEngine,
)
