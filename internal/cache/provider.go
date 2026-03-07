package cache

import "github.com/google/wire"

func ProvideCache() (ICache[string, any], error) {
	return NewCache[string, any]()
}

func ProvideCleanup(cache ICache[string, any]) func() error {
	return func() error {
		if cache == nil {
			return nil
		}
		return cache.Close()
	}
}

var CacheSet = wire.NewSet(ProvideCache)
var CleanupSet = wire.NewSet(ProvideCleanup)
var ProviderSet = wire.NewSet(CacheSet, CleanupSet)
