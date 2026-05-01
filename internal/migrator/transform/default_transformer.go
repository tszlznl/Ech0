// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package transform

import (
	"context"
	"fmt"
	"strings"

	"github.com/lin-snow/ech0/internal/migrator/spec"
)

type DefaultTransformer struct{}

func NewDefaultTransformer() *DefaultTransformer {
	return &DefaultTransformer{}
}

func (t *DefaultTransformer) Transform(_ context.Context, raw spec.RawRecord) (spec.CanonicalRecord, error) {
	title := strings.TrimSpace(toString(raw.Data["title"]))
	content := strings.TrimSpace(toString(raw.Data["content"]))
	if title == "" && content == "" {
		return spec.CanonicalRecord{}, fmt.Errorf("empty title and content")
	}

	return spec.CanonicalRecord{
		SourceID: raw.SourceID,
		Title:    title,
		Content:  content,
		Meta:     raw.Data,
	}, nil
}

func toString(v any) string {
	if v == nil {
		return ""
	}
	switch x := v.(type) {
	case string:
		return x
	case fmt.Stringer:
		return x.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}
