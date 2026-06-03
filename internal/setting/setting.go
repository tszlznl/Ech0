// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package setting 是配置读取的基础设施层：把每个 KV 配置收敛成一份自描述的
// Spec[T]（key + 默认值 + 归一化 + 可选升级迁移），配一个泛型引擎 Get[T] 与启动期
// seeder。它与 internal/kvstore、internal/storage 同层，范式对齐 internal/event
// 的「纯词汇 + 引擎」：
//
//   - Spec[T] 是不可变的「词汇」——声明某配置的 key、如何从 config 构造默认值、
//     读出后如何归一化。所有默认值因此只声明一次（见 registry.go）。
//   - Get[T] 是引擎：读 KV → 反序列化 → 归一化；KV 缺失时回退到 Default()，
//     即便 seeder 未运行也不返回零值/报错。
//   - Seed 在启动期（BeforeStart）把缺失的 key 落库一次（幂等，绝不覆盖用户值），
//     此后各读路径都能命中，Get 不必再承担「读时 seed」的副作用。
//
// 依赖方向只向下：本包仅 import kvstore / config / model / util，绝不 import
// service/handler，也不被 kvstore/config 反向依赖，故可被任意业务层安全引用而不成环。
package setting

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/lin-snow/ech0/internal/kvstore"
)

// Spec 描述单个 KV 配置的读取契约。
type Spec[T any] struct {
	// Key 是该配置在 KV 中的键名。
	Key string
	// Default 构造缺省值（按需读 config.Config() 单例）。
	Default func() T
	// Normalize 在读出/构造默认后就地归一化（可空）；方向恒为 config→value，
	// 绝不反向写回 config。
	Normalize func(*T)
	// Migrate 仅在 seeding 阶段、且 key 尚不存在时尝试从历史数据迁移（可空）。
	// 返回 ok=true 时用其结果取代 Default()，用于平滑升级（如 Passkey 从旧
	// oauth2_setting 搬迁 WebAuthn 字段）。Get 的读路径不走它。
	Migrate func(context.Context, kvstore.Store) (T, bool)
}

// Get 读取并反序列化某配置；KV 缺失（ErrNotFound）时回退到 Default()，并对结果
// 跑一次 Normalize，使「已落库的旧值」与「config 默认」走同一归一化路径。
//
// 后端真实故障（非 ErrNotFound）时仍返回一份可用的归一化默认值，同时上抛该错误：
// 忽略 err 的调用方因此永远拿到 sane value，检查 err 的调用方可据此记录/降级。
func Get[T any](ctx context.Context, kv kvstore.Store, spec Spec[T]) (T, error) {
	raw, err := kv.Get(ctx, spec.Key)
	if err != nil {
		v := spec.Default()
		if spec.Normalize != nil {
			spec.Normalize(&v)
		}
		if errors.Is(err, kvstore.ErrNotFound) {
			return v, nil
		}
		return v, err
	}

	var v T
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return v, err
	}
	if spec.Normalize != nil {
		spec.Normalize(&v)
	}
	return v, nil
}

// Set 归一化 → 序列化 → 落库（upsert）。与 Get 共用同一 Spec，spec 因此成为该配置的
// 双向编解码器。与 Get 对称：它不是裸 upsert，会先跑 spec.Normalize（真正的裸写是底层
// kvstore.Store.Set）。鉴权/校验/副作用属调用方（service 层）职责，不在此原语内。
func Set[T any](ctx context.Context, kv kvstore.Store, spec Spec[T], value T) error {
	if spec.Normalize != nil {
		spec.Normalize(&value)
	}
	buf, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return kv.Set(ctx, spec.Key, string(buf))
}

// seedable 让异构的 Spec[T]（不同 T）能装进同一个 registry 一起 seeding。
type seedable interface {
	seed(ctx context.Context, kv kvstore.Store) error
}

// seed 实现 seedable：仅当 key 不存在时写入（Migrate 命中则用迁移值，否则用
// Default），幂等，绝不覆盖用户已存的值。后端故障（非 ErrNotFound）时冒泡，
// 不在状态不确定时贸然写默认。
func (s Spec[T]) seed(ctx context.Context, kv kvstore.Store) error {
	if _, err := kv.Get(ctx, s.Key); err == nil {
		return nil
	} else if !errors.Is(err, kvstore.ErrNotFound) {
		return err
	}

	var v T
	if s.Migrate != nil {
		if migrated, ok := s.Migrate(ctx, kv); ok {
			v = migrated
		} else {
			v = s.Default()
		}
	} else {
		v = s.Default()
	}
	if s.Normalize != nil {
		s.Normalize(&v)
	}

	buf, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return kv.Set(ctx, s.Key, string(buf))
}

// Seed 把 registry 里所有缺失的配置 key 落库为默认值。由 app 在 BeforeStart 阶段
// 调用一次（DB 已迁移、kvstore 已就绪）。
func Seed(ctx context.Context, kv kvstore.Store) error {
	for _, s := range registry {
		if err := s.seed(ctx, kv); err != nil {
			return err
		}
	}
	return nil
}
