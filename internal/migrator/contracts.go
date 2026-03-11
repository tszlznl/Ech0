package migrator

import "github.com/lin-snow/ech0/internal/migrator/spec"

type (
	ExtractRequest  = spec.ExtractRequest
	ExtractResult   = spec.ExtractResult
	RawRecord       = spec.RawRecord
	CanonicalRecord = spec.CanonicalRecord
	FailedItem      = spec.FailedItem
	LoadResult      = spec.LoadResult
	MigrateRequest  = spec.MigrateRequest
	MigrateProgress = spec.MigrateProgress
	MigrateResult   = spec.MigrateResult
)

type (
	Extractor      = spec.Extractor
	SourceMigrator = spec.SourceMigrator
	Transformer    = spec.Transformer
	Validator      = spec.Validator
	Loader         = spec.Loader
)
