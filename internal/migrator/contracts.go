package migrator

import "github.com/lin-snow/ech0/internal/migrator/spec"

type ExtractRequest = spec.ExtractRequest
type ExtractResult = spec.ExtractResult
type RawRecord = spec.RawRecord
type CanonicalRecord = spec.CanonicalRecord
type FailedItem = spec.FailedItem
type LoadResult = spec.LoadResult
type MigrateRequest = spec.MigrateRequest
type MigrateProgress = spec.MigrateProgress
type MigrateResult = spec.MigrateResult

type Extractor = spec.Extractor
type SourceMigrator = spec.SourceMigrator
type Transformer = spec.Transformer
type Validator = spec.Validator
type Loader = spec.Loader
