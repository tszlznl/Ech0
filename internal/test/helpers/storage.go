// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package helpers

import (
	"testing"

	"github.com/lin-snow/ech0/internal/storage"
)

// NewTestStorage 返回一个仅本地、根目录落在 t.TempDir() 的 storage.Manager，
// 供 file service 测试跑真实的上传/读取/删除（对象存储关闭、无 DB 依赖）。
// 临时目录由 testing 框架在测试结束时自动清理，互不干扰、可并发跑不同包。
func NewTestStorage(t *testing.T) *storage.Manager {
	t.Helper()
	return storage.NewStorageManagerForTest(t.TempDir())
}
