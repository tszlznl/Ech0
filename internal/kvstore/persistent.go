// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package kvstore

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

// Persistent 是持久化键值存储：把统一的 Get/Set/Delete 委托给底层 KeyValue 仓储
// （Backend）。Set 走仓储的 upsert；Get 把底层的 gorm.ErrRecordNotFound 归一化为
// kvstore.ErrNotFound，对上层屏蔽具体持久化实现。gorm 依赖仅存在于此适配器。
//
// 实例化时，持有它的字段按约定命名 durableKV（见包注释）。
type Persistent struct {
	backend Backend
}

var _ Store = (*Persistent)(nil)

// NewPersistent 用给定的底层仓储构造持久化键值存储。
func NewPersistent(backend Backend) *Persistent {
	return &Persistent{backend: backend}
}

// Get 读取键值；底层「记录不存在」统一归一化为 ErrNotFound。
func (s *Persistent) Get(ctx context.Context, key string) (string, error) {
	v, err := s.backend.GetKeyValue(ctx, key)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", ErrNotFound
	}
	return v, err
}

// Set 写入键值（upsert）。
func (s *Persistent) Set(ctx context.Context, key, value string) error {
	return s.backend.AddOrUpdateKeyValue(ctx, key, value)
}

// Delete 删除键。
func (s *Persistent) Delete(ctx context.Context, key string) error {
	return s.backend.DeleteKeyValue(ctx, key)
}
