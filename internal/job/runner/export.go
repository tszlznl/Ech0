// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package runner

import (
	"context"

	"github.com/lin-snow/ech0/internal/event"
	eventbus "github.com/lin-snow/ech0/internal/event/bus"
	"github.com/lin-snow/ech0/internal/job"
	coreMigrator "github.com/lin-snow/ech0/internal/migrator"
	migratorModel "github.com/lin-snow/ech0/internal/model/migrator"
	"github.com/lin-snow/ech0/pkg/busen"
)

// SnapshotExporter 是导出执行端，便于测试解耦（由 migrator.ExportEngine 满足）。
type SnapshotExporter interface {
	Export(ctx context.Context, report func(phase string, snapshot any)) (coreMigrator.ExportOutcome, error)
}

var _ SnapshotExporter = (*coreMigrator.ExportEngine)(nil)

// ExportRunner 把导出执行包成作业 Runner（手动快照异步出口）。导出完成后发布 SystemSnapshot
// 事件（webhook 观察 SystemSnapshot）。
type ExportRunner struct {
	exporter SnapshotExporter
	bus      *busen.Bus
}

func NewExportRunner(exporter *coreMigrator.ExportEngine, busProvider func() *busen.Bus) *ExportRunner {
	return &ExportRunner{exporter: exporter, bus: busProvider()}
}

func (r *ExportRunner) Run(ctx context.Context, _ migratorModel.ExportPayload, report job.ReportFunc) (any, error) {
	outcome, err := r.exporter.Export(ctx, report)
	if err != nil {
		return nil, err
	}

	eventbus.Notify(ctx, r.bus, event.SystemSnapshot{Info: "System manual snapshot completed"})

	return outcome, nil
}
