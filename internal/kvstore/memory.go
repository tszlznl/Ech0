// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package kvstore

import (
	"context"
	"sync"
)

// Memory 是进程内键值存储：基于 map + RWMutex，确定性、零外部依赖。它本身即
// 权威数据源（非缓存），不带 TTL / 驱逐，结果可预测——主要用作单元测试中 Store
// 的替身，也可承载「丢失可接受」的临时数据。
//
// 实例化时，持有它的字段按约定命名 ephemeralKV（见包注释）。
type Memory struct {
	mu sync.RWMutex
	m  map[string]string
}

var _ Store = (*Memory)(nil)

// NewMemory 创建空的内存键值存储。
func NewMemory() *Memory {
	return &Memory{m: make(map[string]string)}
}

// Get 返回键对应的值；键不存在时返回 ErrNotFound。
func (s *Memory) Get(_ context.Context, key string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.m[key]
	if !ok {
		return "", ErrNotFound
	}
	return v, nil
}

// Set 写入键值（upsert）。
func (s *Memory) Set(_ context.Context, key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[key] = value
	return nil
}

// Delete 删除键（键不存在时为 no-op）。
func (s *Memory) Delete(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, key)
	return nil
}
