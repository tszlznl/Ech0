// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package runner

import (
	"context"

	"github.com/lin-snow/ech0/internal/job"
	migrationModel "github.com/lin-snow/ech0/internal/model/migration"
	migratorService "github.com/lin-snow/ech0/internal/service/migrator"
)

// MigrationImporter 是迁移导入执行端，便于测试解耦（由 migratorService.Importer 满足）。
type MigrationImporter interface {
	Import(ctx context.Context, payload migrationModel.MigrationPayload, report func(phase string, snapshot any)) (any, error)
}

var _ MigrationImporter = (*migratorService.Importer)(nil)

// MigrationRunner 把迁移导入包成作业 Runner。
type MigrationRunner struct {
	importer MigrationImporter
}

func NewMigrationRunner(importer *migratorService.Importer) *MigrationRunner {
	return &MigrationRunner{importer: importer}
}

func (r *MigrationRunner) Run(ctx context.Context, p migrationModel.MigrationPayload, report job.ReportFunc) (any, error) {
	return r.importer.Import(ctx, p, report)
}
