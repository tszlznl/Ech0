package cache

import (
	"testing"
	"time"
)

type closeSpyCache struct {
	closed bool
}

func (c *closeSpyCache) Set(string, any, int64) bool { return true }
func (c *closeSpyCache) SetWithTTL(string, any, int64, time.Duration) bool {
	return true
}
func (c *closeSpyCache) Get(string) (any, bool, error) { return nil, false, nil }
func (c *closeSpyCache) Delete(string)                 {}
func (c *closeSpyCache) Close() error {
	c.closed = true
	return nil
}

func TestCacheFactoryCleanup(t *testing.T) {
	spy := &closeSpyCache{}
	f := &CacheFactory{cache: spy}

	if err := f.Cleanup(); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}
	if !spy.closed {
		t.Fatalf("expected cache close to be called")
	}
}

func TestCacheFactoryCleanupNilFactory(t *testing.T) {
	var f *CacheFactory
	if err := f.Cleanup(); err != nil {
		t.Fatalf("cleanup should not fail for nil factory: %v", err)
	}
}
