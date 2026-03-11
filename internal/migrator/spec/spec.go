package spec

import "context"

type ExtractRequest struct {
	SourcePayload map[string]any
	Checkpoint    int64
	BatchSize     int
}

type ExtractResult struct {
	Records        []RawRecord
	NextCheckpoint int64
	HasMore        bool
	TotalHint      int64
}

type RawRecord struct {
	SourceID string
	Data     map[string]any
}

type CanonicalRecord struct {
	SourceID string
	Title    string
	Content  string
	Meta     map[string]any
}

type FailedItem struct {
	SourceID string `json:"source_id"`
	Reason   string `json:"reason"`
}

type LoadResult struct {
	Loaded int64
	Failed []FailedItem
}

type MigrateRequest struct {
	SourcePayload  map[string]any
	UpdateProgress func(progress MigrateProgress)
}

type MigrateProgress struct {
	CurrentPhase string
	Processed    int64
	Total        int64
	SuccessCount int64
	FailCount    int64
	ErrorSummary string
}

type MigrateResult struct {
	Processed    int64
	Total        int64
	SuccessCount int64
	FailCount    int64
	ErrorSummary string
	JobID        string
	Report       map[string]any
}

type Extractor interface {
	Extract(ctx context.Context, req ExtractRequest) (ExtractResult, error)
}

type SourceMigrator interface {
	Migrate(ctx context.Context, req MigrateRequest) (MigrateResult, error)
}

type Transformer interface {
	Transform(ctx context.Context, raw RawRecord) (CanonicalRecord, error)
}

type Validator interface {
	Validate(ctx context.Context, record CanonicalRecord) error
}

type Loader interface {
	Load(ctx context.Context, records []CanonicalRecord) (LoadResult, error)
}
