// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package runner 实现各领域的作业 Runner，连接通用 job 框架与领域 service。
package runner

import (
	"context"

	"github.com/lin-snow/ech0/internal/job"
	embeddingService "github.com/lin-snow/ech0/internal/service/embedding"
)

// ReindexPayload 无输入（全量重建）。
type ReindexPayload struct{}

// ReindexRunner 把 EmbeddingService.Backfill 包成作业 Runner。
type ReindexRunner struct {
	svc embeddingService.Service
}

func NewReindexRunner(svc embeddingService.Service) *ReindexRunner {
	return &ReindexRunner{svc: svc}
}

// Run 跑 Backfill，每页结束上报累计计数；终态 result 为 BackfillResult。
func (r *ReindexRunner) Run(ctx context.Context, _ ReindexPayload, report job.ReportFunc) (any, error) {
	res, err := r.svc.Backfill(ctx, func(progress embeddingService.BackfillResult) {
		report("indexing", progress)
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}
