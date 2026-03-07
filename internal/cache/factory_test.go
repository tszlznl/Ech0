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

func TestProvideCleanup(t *testing.T) {
	spy := &closeSpyCache{}
	cleanup := ProvideCleanup(spy)

	if err := cleanup(); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}
	if !spy.closed {
		t.Fatalf("expected cache close to be called")
	}
}

func TestProvideCleanupNilCache(t *testing.T) {
	cleanup := ProvideCleanup(nil)
	if err := cleanup(); err != nil {
		t.Fatalf("cleanup should not fail for nil cache: %v", err)
	}
}
