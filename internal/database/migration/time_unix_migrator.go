// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package migration

import (
	"fmt"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	"gorm.io/gorm"
)

type storageTimeUnixMigrator struct{}

func NewStorageTimeUnixMigrator() Migrator {
	return &storageTimeUnixMigrator{}
}

func (m *storageTimeUnixMigrator) Name() string {
	return "storage_time_unix_migrator"
}

func (m *storageTimeUnixMigrator) Key() string {
	return commonModel.StorageTimeUnixMigratedKey
}

func (m *storageTimeUnixMigrator) CanRerun() bool {
	return false
}

func (m *storageTimeUnixMigrator) Migrate(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for _, plan := range StorageTimeColumnPlans() {
			updateSQL := fmt.Sprintf(
				"UPDATE %s SET %s = CAST(strftime('%%s', %s) AS INTEGER) WHERE %s IS NOT NULL AND typeof(%s) = 'text'",
				plan.Table,
				plan.Column,
				plan.Column,
				plan.Column,
				plan.Column,
			)
			if err := tx.Exec(updateSQL).Error; err != nil {
				return fmt.Errorf("convert %s.%s to unix failed: %w", plan.Table, plan.Column, err)
			}
		}
		return nil
	})
}
