package validate

import (
	"context"
	"errors"
	"strings"

	"github.com/lin-snow/ech0/internal/migrator/spec"
)

type DefaultValidator struct{}

func NewDefaultValidator() *DefaultValidator {
	return &DefaultValidator{}
}

func (v *DefaultValidator) Validate(_ context.Context, record spec.CanonicalRecord) error {
	if strings.TrimSpace(record.SourceID) == "" {
		return errors.New("source id is empty")
	}
	if strings.TrimSpace(record.Title) == "" && strings.TrimSpace(record.Content) == "" {
		return errors.New("title and content are both empty")
	}
	return nil
}
