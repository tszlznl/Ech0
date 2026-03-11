package uuid

import (
	"fmt"

	"github.com/google/uuid"
)

// NewV7 returns a string UUIDv7 identifier.
func NewV7() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	return id.String(), nil
}

// MustNewV7 returns a UUIDv7 string and panics on error.
func MustNewV7() string {
	id, err := NewV7()
	if err != nil {
		panic(fmt.Errorf("generate uuid v7: %w", err))
	}
	return id
}

// IsValid reports whether s is a valid UUID string.
func IsValid(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}
