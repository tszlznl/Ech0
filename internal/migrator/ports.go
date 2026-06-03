// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package migrator

import (
	"github.com/lin-snow/ech0/internal/cache"
	"github.com/lin-snow/ech0/internal/kvstore"
	"github.com/lin-snow/ech0/internal/storage"
)

// 这些端口是引擎执行体(Importer/Exporter)所需的基础设施依赖,由 DI 注入。放在核心
// 引擎包(而非 service 层),是因为导入/导出的执行逻辑本身就属于 Migrator 引擎;service
// 层只做 auth + 作业生命周期 + DTO 转发。
//
// 持有 KVStore 的字段按 kvstore 包约定命名 durableKV(数据需活过重启)。
type (
	KVStore        = kvstore.Store
	StorageManager = *storage.Manager
	AppCache       = cache.ICache[string, any]
)
