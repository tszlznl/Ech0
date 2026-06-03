// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package kvstore 提供统一的键值存储抽象：一套 Get/Set/Delete 契约（Store），
// 配两个实现——Memory（进程内 map，确定性、零依赖，主要用作测试替身，也可
// 承载「丢失可接受」的临时数据）与 Persistent（持久化，委托底层 KeyValue 仓储）。
// 各业务包需要 KV 时统一依赖 kvstore.Store，而非耦合具体仓储，便于替换与测试。
//
// 设计与 internal/storage 对齐：kvstore 自己定义 Backend 接口描述「持久化实现
// 需要从仓储拿到的能力」，由 Wire 绑定到 repository/keyvalue，kvstore 因此不
// 反向 import repository 层。
//
// # 字段命名约定（实例化此抽象时的官方推荐）
//
// 持有 Store 的字段按「数据是否活过重启」这一**契约**命名，而非按实现类命名：
//   - durableKV   —— 由 Persistent 支撑、需持久化的存储（配置、需长期保存的数据）；
//   - ephemeralKV —— 由 Memory 支撑、丢失可接受的存储（会话、限流计数、缓存类）。
//
// 单一存储时也用 durableKV / ephemeralKV，不要写成 persistentKV / memKV——后者把
// 实现机制泄露进消费者。当一个服务同时依赖两种存储时，两个字段名即自解释其契约。
package kvstore

import (
	"context"
	"errors"
)

// ErrNotFound 表示键不存在。两个实现统一返回它，调用方可用 errors.Is 判定。
var ErrNotFound = errors.New("kvstore: key not found")

// Store 是统一的键值存储契约。Set 为 upsert 语义（存在即覆盖，不存在即新增）。
type Store interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string) error
	Delete(ctx context.Context, key string) error
}

// Backend 是 Persistent 委托的底层仓储能力（方法名对齐 repository/keyvalue，
// 使其结构性满足而无需改动仓储）。由 Wire 绑定到具体的 KeyValue 仓储。
type Backend interface {
	GetKeyValue(ctx context.Context, key string) (string, error)
	AddOrUpdateKeyValue(ctx context.Context, key, value string) error
	DeleteKeyValue(ctx context.Context, key string) error
}
