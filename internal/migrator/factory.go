package migrator

import (
	"fmt"

	ech0v3Extractor "github.com/lin-snow/ech0/internal/migrator/extractor/ech0v3"
	memosExtractor "github.com/lin-snow/ech0/internal/migrator/extractor/memos"
	"github.com/lin-snow/ech0/internal/migrator/load"
	"github.com/lin-snow/ech0/internal/migrator/transform"
	"github.com/lin-snow/ech0/internal/migrator/validate"
	migrationModel "github.com/lin-snow/ech0/internal/model/migration"
)

func BuildRunner(sourceType string, createdBy string) (*Runner, error) {
	var extractor Extractor
	switch sourceType {
	case migrationModel.MigrationSourceMemos:
		extractor = memosExtractor.NewExtractor()
	case migrationModel.MigrationSourceEch0V3:
		extractor = ech0v3Extractor.NewExtractor()
	default:
		return nil, fmt.Errorf("unsupported migration source type: %s", sourceType)
	}

	return NewRunner(
		extractor,
		transform.NewDefaultTransformer(),
		validate.NewDefaultValidator(),
		load.NewEchoLoader(createdBy),
	), nil
}
