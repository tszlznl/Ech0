// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package migration

import (
	"fmt"
	"regexp"
	"strings"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	"gorm.io/gorm"
)

type storageTimeSchemaRebuildMigrator struct{}

func NewStorageTimeSchemaRebuildMigrator() Migrator {
	return &storageTimeSchemaRebuildMigrator{}
}

func (m *storageTimeSchemaRebuildMigrator) Name() string {
	return "storage_time_schema_rebuild_migrator"
}

func (m *storageTimeSchemaRebuildMigrator) Key() string {
	return commonModel.StorageTimeSchemaRebuiltKey
}

func (m *storageTimeSchemaRebuildMigrator) CanRerun() bool {
	return false
}

func (m *storageTimeSchemaRebuildMigrator) Migrate(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("PRAGMA foreign_keys = OFF").Error; err != nil {
			return fmt.Errorf("disable foreign keys failed: %w", err)
		}
		defer tx.Exec("PRAGMA foreign_keys = ON")

		tableColumns := StorageTimeColumnsByTable()
		for table, columns := range tableColumns {
			needsRebuild, err := tableNeedsRebuild(tx, table, columns)
			if err != nil {
				return err
			}
			if !needsRebuild {
				continue
			}

			if err := rebuildTableWithIntegerColumns(tx, table, columns); err != nil {
				return err
			}
		}
		return nil
	})
}

func tableNeedsRebuild(db *gorm.DB, table string, columns []string) (bool, error) {
	columnTypeMap, _, err := loadTableInfo(db, table)
	if err != nil {
		return false, err
	}
	for _, column := range columns {
		typeName := strings.TrimSpace(columnTypeMap[column])
		if !strings.Contains(strings.ToUpper(typeName), "INT") {
			return true, nil
		}
	}
	return false, nil
}

func rebuildTableWithIntegerColumns(db *gorm.DB, table string, columns []string) error {
	var createSQL string
	if err := db.Raw("SELECT sql FROM sqlite_master WHERE type='table' AND name = ?", table).Scan(&createSQL).Error; err != nil {
		return fmt.Errorf("load create sql for %s failed: %w", table, err)
	}
	if strings.TrimSpace(createSQL) == "" {
		return fmt.Errorf("empty create sql for table %s", table)
	}

	indexSQLs, triggerSQLs, err := loadDDLDependencies(db, table)
	if err != nil {
		return err
	}

	tempTable := fmt.Sprintf("%s__time_rebuild", table)
	tempCreateSQL, err := buildTempCreateSQL(createSQL, table, tempTable, columns)
	if err != nil {
		return err
	}
	if err := db.Exec(tempCreateSQL).Error; err != nil {
		return fmt.Errorf("create temp table %s failed: %w", tempTable, err)
	}

	_, orderedColumns, err := loadTableInfo(db, table)
	if err != nil {
		return err
	}
	columnList := quoteColumnList(orderedColumns)
	copySQL := fmt.Sprintf("INSERT INTO \"%s\" (%s) SELECT %s FROM \"%s\"", tempTable, columnList, columnList, table)
	if err := db.Exec(copySQL).Error; err != nil {
		return fmt.Errorf("copy data %s -> %s failed: %w", table, tempTable, err)
	}

	if err := db.Exec(fmt.Sprintf("DROP TABLE \"%s\"", table)).Error; err != nil {
		return fmt.Errorf("drop old table %s failed: %w", table, err)
	}
	if err := db.Exec(fmt.Sprintf("ALTER TABLE \"%s\" RENAME TO \"%s\"", tempTable, table)).Error; err != nil {
		return fmt.Errorf("rename temp table %s back to %s failed: %w", tempTable, table, err)
	}

	for _, ddl := range append(indexSQLs, triggerSQLs...) {
		if strings.TrimSpace(ddl) == "" {
			continue
		}
		if err := db.Exec(ddl).Error; err != nil {
			return fmt.Errorf("recreate ddl for %s failed: %w", table, err)
		}
	}
	return nil
}

func loadDDLDependencies(db *gorm.DB, table string) ([]string, []string, error) {
	var indexSQLs []string
	if err := db.Raw("SELECT sql FROM sqlite_master WHERE type='index' AND tbl_name=? AND sql IS NOT NULL", table).Scan(&indexSQLs).Error; err != nil {
		return nil, nil, fmt.Errorf("load index ddl for %s failed: %w", table, err)
	}
	var triggerSQLs []string
	if err := db.Raw("SELECT sql FROM sqlite_master WHERE type='trigger' AND tbl_name=? AND sql IS NOT NULL", table).Scan(&triggerSQLs).Error; err != nil {
		return nil, nil, fmt.Errorf("load trigger ddl for %s failed: %w", table, err)
	}
	return indexSQLs, triggerSQLs, nil
}

func buildTempCreateSQL(createSQL, table, tempTable string, columns []string) (string, error) {
	tablePattern := regexp.MustCompile(fmt.Sprintf(`(?i)^CREATE TABLE\s+(IF NOT EXISTS\s+)?([`+"`"+`"\[]?%s[`+"`"+`"\]]?)`, regexp.QuoteMeta(table)))
	tempSQL := tablePattern.ReplaceAllString(createSQL, fmt.Sprintf("CREATE TABLE ${1}\"%s\"", tempTable))
	if tempSQL == createSQL {
		return "", fmt.Errorf("rewrite create sql table name for %s failed", table)
	}

	for _, column := range columns {
		colPattern := regexp.MustCompile(fmt.Sprintf(`(?i)(["`+"`"+`\[]?%s["`+"`"+`\]]?\s+)([A-Z]+(?:\s*\(\s*\d+\s*\))?)`, regexp.QuoteMeta(column)))
		tempSQL = colPattern.ReplaceAllString(tempSQL, `${1}INTEGER`)
	}
	return tempSQL, nil
}

func loadTableInfo(db *gorm.DB, table string) (map[string]string, []string, error) {
	var pragmaRows []struct {
		Name string `gorm:"column:name"`
		Type string `gorm:"column:type"`
	}
	if err := db.Raw(fmt.Sprintf("PRAGMA table_info(\"%s\")", table)).Scan(&pragmaRows).Error; err != nil {
		return nil, nil, fmt.Errorf("load table_info for %s failed: %w", table, err)
	}
	typeMap := make(map[string]string, len(pragmaRows))
	ordered := make([]string, 0, len(pragmaRows))
	for _, row := range pragmaRows {
		typeMap[row.Name] = row.Type
		ordered = append(ordered, row.Name)
	}
	return typeMap, ordered, nil
}

func quoteColumnList(columns []string) string {
	quoted := make([]string, 0, len(columns))
	for _, column := range columns {
		quoted = append(quoted, fmt.Sprintf("\"%s\"", column))
	}
	return strings.Join(quoted, ",")
}
