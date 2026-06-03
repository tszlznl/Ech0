// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package migrator

import (
	"fmt"

	fsExporter "github.com/lin-snow/ech0/internal/migrator/exporter/fs"
	s3Exporter "github.com/lin-snow/ech0/internal/migrator/exporter/s3"
	ech0Importer "github.com/lin-snow/ech0/internal/migrator/importer/ech0"
	memosImporter "github.com/lin-snow/ech0/internal/migrator/importer/memos"
	"github.com/lin-snow/ech0/internal/migrator/spec"
	migratorModel "github.com/lin-snow/ech0/internal/model/migrator"
)

// BuildImporter 按来源选导入适配器(ech0 / memos),与 BuildExporter 对称。
func BuildImporter(source string) (spec.Importer, error) {
	switch source {
	case migratorModel.MigrationSourceEch0:
		return ech0Importer.New(), nil
	case migratorModel.MigrationSourceMemos:
		return memosImporter.New(), nil
	default:
		return nil, fmt.Errorf("unsupported import source: %s", source)
	}
}

// BuildExporter 按目的地选导出适配器(fs / s3),与 BuildImporter 对称。s3 需 storageManager 取配置。
func BuildExporter(dest string, storageManager StorageManager) (spec.Exporter, error) {
	switch dest {
	case migratorModel.ExportDestFS:
		return fsExporter.New(), nil
	case migratorModel.ExportDestS3:
		return s3Exporter.New(storageManager), nil
	default:
		return nil, fmt.Errorf("unsupported export destination: %s", dest)
	}
}
