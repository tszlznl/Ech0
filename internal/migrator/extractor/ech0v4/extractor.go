package ech0v4

import (
	"context"

	"github.com/lin-snow/ech0/internal/migrator/spec"
)

type Extractor struct{}

func NewExtractor() *Extractor {
	return &Extractor{}
}

func (e *Extractor) Extract(_ context.Context, req spec.ExtractRequest) (spec.ExtractResult, error) {
	// ech0_v4 导入逻辑由后续迭代补齐；当前先占位保证任务链路可用。
	return spec.ExtractResult{
		Records:        []spec.RawRecord{},
		NextCheckpoint: req.Checkpoint,
		HasMore:        false,
		TotalHint:      0,
	}, nil
}

func (e *Extractor) Migrate(_ context.Context, req spec.MigrateRequest) (spec.MigrateResult, error) {
	if req.UpdateProgress != nil {
		req.UpdateProgress(spec.MigrateProgress{
			CurrentPhase: specPhaseLoading,
			Processed:    0,
			Total:        0,
			SuccessCount: 0,
			FailCount:    0,
		})
	}
	return spec.MigrateResult{
		Processed:    0,
		Total:        0,
		SuccessCount: 0,
		FailCount:    0,
		ErrorSummary: "",
	}, nil
}

const specPhaseLoading = "loading"
