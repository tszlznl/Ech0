package cache

import "github.com/google/wire"

func ProvideCache(factory *CacheFactory) ICache[string, any] {
	return factory.Cache()
}

func ProvideCleanup(factory *CacheFactory) func() error {
	return factory.Cleanup
}

var FactorySet = wire.NewSet(NewCacheFactory)
var CacheSet = wire.NewSet(ProvideCache)
var CleanupSet = wire.NewSet(ProvideCleanup)
var ProviderSet = wire.NewSet(FactorySet, CacheSet, CleanupSet)
