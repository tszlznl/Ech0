// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package runner

import (
	"context"

	contracts "github.com/lin-snow/ech0/internal/event/contracts"
	publisher "github.com/lin-snow/ech0/internal/event/publisher"
	"github.com/lin-snow/ech0/internal/job"
	coreMigrator "github.com/lin-snow/ech0/internal/migrator"
	migratorModel "github.com/lin-snow/ech0/internal/model/migrator"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"go.uber.org/zap"
)

// SnapshotExporter 是导出执行端，便于测试解耦（由 migrator.ExportEngine 满足）。
type SnapshotExporter interface {
	Export(ctx context.Context, report func(phase string, snapshot any)) (coreMigrator.ExportOutcome, error)
}

var _ SnapshotExporter = (*coreMigrator.ExportEngine)(nil)

// ExportRunner 把导出执行包成作业 Runner（手动快照异步出口）。导出完成后发布 SystemSnapshot
// 事件（webhook 订阅 TopicSystemSnapshot）。
type ExportRunner struct {
	exporter  SnapshotExporter
	publisher *publisher.Publisher
}

func NewExportRunner(exporter *coreMigrator.ExportEngine, publisher *publisher.Publisher) *ExportRunner {
	return &ExportRunner{exporter: exporter, publisher: publisher}
}

func (r *ExportRunner) Run(ctx context.Context, _ migratorModel.ExportPayload, report job.ReportFunc) (any, error) {
	outcome, err := r.exporter.Export(ctx, report)
	if err != nil {
		return nil, err
	}

	if r.publisher != nil {
		if pubErr := r.publisher.SystemSnapshot(
			ctx,
			contracts.SystemSnapshotEvent{Info: "System manual snapshot completed"},
		); pubErr != nil {
			logUtil.GetLogger().Error("Failed to publish system snapshot completed event", zap.Error(pubErr))
		}
	}

	return outcome, nil
}
