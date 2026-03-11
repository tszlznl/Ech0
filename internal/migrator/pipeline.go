package migrator

import (
	"context"
	"errors"
)

type Runner struct {
	extractor   Extractor
	transformer Transformer
	validator   Validator
	loader      Loader
}

func NewRunner(
	extractor Extractor,
	transformer Transformer,
	validator Validator,
	loader Loader,
) *Runner {
	return &Runner{
		extractor:   extractor,
		transformer: transformer,
		validator:   validator,
		loader:      loader,
	}
}

type BatchOutcome struct {
	TotalHint      int64
	NextCheckpoint int64
	HasMore        bool
	Loaded         int64
	Failed         []FailedItem
}

func (r *Runner) RunBatch(ctx context.Context, req ExtractRequest) (BatchOutcome, error) {
	if r.extractor == nil || r.transformer == nil || r.validator == nil || r.loader == nil {
		return BatchOutcome{}, errors.New("migrator pipeline is not fully configured")
	}

	extractResult, err := r.extractor.Extract(ctx, req)
	if err != nil {
		return BatchOutcome{}, err
	}

	canonicalRecords := make([]CanonicalRecord, 0, len(extractResult.Records))
	failed := make([]FailedItem, 0)
	for _, raw := range extractResult.Records {
		record, transformErr := r.transformer.Transform(ctx, raw)
		if transformErr != nil {
			failed = append(failed, FailedItem{SourceID: raw.SourceID, Reason: transformErr.Error()})
			continue
		}
		if validateErr := r.validator.Validate(ctx, record); validateErr != nil {
			failed = append(failed, FailedItem{SourceID: raw.SourceID, Reason: validateErr.Error()})
			continue
		}
		canonicalRecords = append(canonicalRecords, record)
	}

	loadResult, err := r.loader.Load(ctx, canonicalRecords)
	if err != nil {
		return BatchOutcome{}, err
	}

	allFailed := append(failed, loadResult.Failed...)
	return BatchOutcome{
		TotalHint:      extractResult.TotalHint,
		NextCheckpoint: extractResult.NextCheckpoint,
		HasMore:        extractResult.HasMore,
		Loaded:         loadResult.Loaded,
		Failed:         allFailed,
	}, nil
}
