// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package helpers 提供后端单元测试共享的脚手架：内存数据库、身份上下文、配置覆写、
// 响应封套解析与常用 fixture。仅被各包的 _test.go 导入，不会进入生产二进制。用法：
//
//	import "github.com/lin-snow/ech0/internal/test/helpers"
//	db := helpers.NewTestDB(t)
package helpers

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	"github.com/lin-snow/ech0/internal/database"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	dbSeq       atomic.Int64
	vecAutoOnce sync.Once
)

// NewTestDB 返回一个隔离的内存 SQLite *gorm.DB，已建好全部表，并设为全局 database.GetDB() 源。
//
// 每个调用使用唯一命名的内存库（mode=memory&cache=shared：同名 DSN 的多条连接共享同一库，
// 不同测试彼此隔离）。t.Cleanup 负责关闭连接并还原此前的全局 DB。
//
// 注意：因依赖全局单例 database.SetDB，使用本 harness 的测试不要 t.Parallel()。
func NewTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	return newTestDB(t)
}

// NewTestDBWithVec 在 NewTestDB 基础上注册 sqlite-vec 进程级自动扩展（vec0 虚表），
// 供 embedding 仓储类测试使用。Auto 只注册一次。
func NewTestDBWithVec(t *testing.T) *gorm.DB {
	t.Helper()
	vecAutoOnce.Do(sqlite_vec.Auto)
	return newTestDB(t)
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:ech0test_%s_%d?mode=memory&cache=shared", dsnSafe(t.Name()), dbSeq.Add(1))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("helpers: open in-memory sqlite: %v", err)
	}

	prev := currentDB()
	database.SetDB(db)
	t.Cleanup(func() {
		if sqlDB, derr := db.DB(); derr == nil {
			_ = sqlDB.Close()
		}
		database.SetDB(prev) // 还原前一个（可能为 nil）
	})

	if err := database.MigrateDB(); err != nil {
		t.Fatalf("helpers: migrate in-memory db: %v", err)
	}
	return db
}

// currentDB 安全读取当前全局 DB：从未初始化时 database.GetDB() 会 panic，这里吞掉并返回 nil。
func currentDB() (db *gorm.DB) {
	defer func() { _ = recover() }()
	return database.GetDB()
}

// dsnSafe 把子测试名里的 '/'、空格等转成 DSN 安全字符。
func dsnSafe(name string) string {
	return strings.NewReplacer("/", "_", " ", "_", "#", "_", ":", "_").Replace(name)
}
