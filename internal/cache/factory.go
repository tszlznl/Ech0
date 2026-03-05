package cache

// CacheFactory 是一个工厂类，用于创建和管理不同类型的缓存实例
type CacheFactory struct {
	cache ICache[string, any] // 通用缓存
}

// NewCacheFactory 创建一个新的 CacheFactory 实例，并初始化所需的缓存
func NewCacheFactory() *CacheFactory {
	cache, err := NewCache[string, any]()
	if err != nil {
		panic(err)
	}

	return &CacheFactory{
		cache: cache,
	}
}

// Cache 返回缓存实例
func (f *CacheFactory) Cache() ICache[string, any] {
	return f.cache
}

// Cleanup 释放缓存资源
func (f *CacheFactory) Cleanup() error {
	if f == nil || f.cache == nil {
		return nil
	}
	return f.cache.Close()
}
