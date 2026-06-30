// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package helpers

import (
	"sync"
	"time"

	"github.com/lin-snow/ech0/internal/cache"
)

// NewTestCache 返回一个确定性的内存 ICache（map + mutex），用于需要 cache.ICache 的
// repository 测试（echo/user/keyvalue/auth）。
//
// 与生产用的 Ristretto 不同：Set 同步且立即可见、不做 TTL 过期与容量淘汰，从而避免
// Ristretto 异步写入导致的 read-after-write flaky，让缓存行为在测试中完全可预测。
func NewTestCache() cache.ICache[string, any] {
	return &testCache{m: make(map[string]any)}
}

type testCache struct {
	mu sync.RWMutex
	m  map[string]any
}

func (c *testCache) Set(key string, value any, _ int64) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[key] = value
	return true
}

func (c *testCache) SetWithTTL(key string, value any, cost int64, _ time.Duration) bool {
	return c.Set(key, value, cost)
}

func (c *testCache) Get(key string) (any, bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.m[key]
	return v, ok, nil
}

func (c *testCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.m, key)
}

func (c *testCache) Close() error { return nil }
