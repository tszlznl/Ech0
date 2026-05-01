// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package migration

import (
	"errors"
	"fmt"
	"strings"
	"time"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Migrator 定义数据库启动后需要执行的迁移任务接口。
type Migrator interface {
	Name() string
	Key() string
	CanRerun() bool
	Migrate(db *gorm.DB) error
}

type migrateOptions struct {
	migrators   []Migrator
	stopOnError bool
}

// Option 用于按需扩展 migration 执行行为。
type Option func(*migrateOptions)

// WithMigrators 追加要执行的迁移器。
func WithMigrators(migrators ...Migrator) Option {
	return func(opts *migrateOptions) {
		opts.migrators = append(opts.migrators, migrators...)
	}
}

// WithStopOnError 配置遇到首个迁移错误时立即停止。
func WithStopOnError() Option {
	return func(opts *migrateOptions) {
		opts.stopOnError = true
	}
}

func defaultOptions() migrateOptions {
	return migrateOptions{
		migrators:   make([]Migrator, 0),
		stopOnError: false,
	}
}

// Migrate 是 migration 子包统一入口，按顺序执行迁移器集合。
func Migrate(db *gorm.DB, optionFns ...Option) {
	if db == nil {
		logUtil.Warn("database migration skipped: db is nil", zap.String("module", "database"))
		return
	}

	opts := defaultOptions()
	for _, fn := range optionFns {
		if fn != nil {
			fn(&opts)
		}
	}

	for _, migrator := range opts.migrators {
		markerKey := strings.TrimSpace(migrator.Key())
		if markerKey != "" && !migrator.CanRerun() {
			done, err := isMigratorDone(db, markerKey)
			if err != nil {
				logUtil.Warn(
					"database migrator state check failed",
					zap.String("module", "database"),
					zap.String("migrator", migrator.Name()),
					zap.String("marker_key", markerKey),
					zap.Error(err),
				)
				if opts.stopOnError {
					return
				}
				continue
			}
			if done {
				logUtil.Info(
					"database migrator skipped (already done)",
					zap.String("module", "database"),
					zap.String("migrator", migrator.Name()),
					zap.String("marker_key", markerKey),
				)
				continue
			}
		}

		if err := migrator.Migrate(db); err != nil {
			logUtil.Warn(
				"database migrator failed",
				zap.String("module", "database"),
				zap.String("migrator", migrator.Name()),
				zap.Error(err),
			)
			if opts.stopOnError {
				return
			}
			continue
		}

		if markerKey != "" {
			if err := markMigratorDone(db, markerKey, migrator.Name()); err != nil {
				logUtil.Warn(
					"database migrator mark done failed",
					zap.String("module", "database"),
					zap.String("migrator", migrator.Name()),
					zap.String("marker_key", markerKey),
					zap.Error(err),
				)
				if opts.stopOnError {
					return
				}
			}
		}
	}
}

func isMigratorDone(db *gorm.DB, markerKey string) (bool, error) {
	var marker commonModel.KeyValue
	err := db.Where("key = ?", markerKey).First(&marker).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(marker.Value) != "", nil
}

func markMigratorDone(db *gorm.DB, markerKey, migratorName string) error {
	marker := commonModel.KeyValue{
		Key: markerKey,
		Value: fmt.Sprintf(
			"done_at=%s;migrator=%s",
			time.Now().UTC().Format(time.RFC3339),
			strings.TrimSpace(migratorName),
		),
	}
	return db.Save(&marker).Error
}
