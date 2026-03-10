package migrator

import "github.com/lin-snow/ech0/internal/migrator/spec"

type ExtractRequest = spec.ExtractRequest
type ExtractResult = spec.ExtractResult
type RawRecord = spec.RawRecord
type CanonicalRecord = spec.CanonicalRecord
type FailedItem = spec.FailedItem
type LoadResult = spec.LoadResult

type Extractor = spec.Extractor
type Transformer = spec.Transformer
type Validator = spec.Validator
type Loader = spec.Loader
