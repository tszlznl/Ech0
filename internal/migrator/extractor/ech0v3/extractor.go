package ech0v3

import (
	"context"
	"fmt"

	"github.com/lin-snow/ech0/internal/migrator/spec"
)

type Extractor struct{}

func NewExtractor() *Extractor {
	return &Extractor{}
}

func (e *Extractor) Extract(_ context.Context, req spec.ExtractRequest) (spec.ExtractResult, error) {
	items, ok := req.SourcePayload["items"].([]any)
	if !ok || len(items) == 0 {
		return spec.ExtractResult{
			Records:        []spec.RawRecord{},
			NextCheckpoint: req.Checkpoint,
			HasMore:        false,
			TotalHint:      0,
		}, nil
	}

	start := int(req.Checkpoint)
	if start < 0 {
		start = 0
	}
	if start >= len(items) {
		return spec.ExtractResult{
			Records:        []spec.RawRecord{},
			NextCheckpoint: int64(len(items)),
			HasMore:        false,
			TotalHint:      int64(len(items)),
		}, nil
	}

	batchSize := req.BatchSize
	if batchSize <= 0 {
		batchSize = 100
	}
	end := start + batchSize
	if end > len(items) {
		end = len(items)
	}

	records := make([]spec.RawRecord, 0, end-start)
	for i := start; i < end; i++ {
		obj, ok := items[i].(map[string]any)
		if !ok {
			return spec.ExtractResult{}, fmt.Errorf("ech0 v3 item at index %d is not object", i)
		}
		if _, exists := obj["content"]; !exists {
			if text, ok := obj["text"]; ok {
				obj["content"] = text
			}
		}
		sourceID := fmt.Sprintf("%v", obj["id"])
		if sourceID == "<nil>" {
			sourceID = fmt.Sprintf("ech0v3-%d", i)
		}
		records = append(records, spec.RawRecord{
			SourceID: sourceID,
			Data:     obj,
		})
	}

	return spec.ExtractResult{
		Records:        records,
		NextCheckpoint: int64(end),
		HasMore:        end < len(items),
		TotalHint:      int64(len(items)),
	}, nil
}
