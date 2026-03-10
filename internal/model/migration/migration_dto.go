package model

type CreateMigrationJobRequest struct {
	SourceType    string         `json:"source_type" binding:"required"`
	SourceVersion string         `json:"source_version"`
	SourcePayload map[string]any `json:"source_payload"`
}

type MigrationJobDTO struct {
	ID             string `json:"id"`
	SourceType     string `json:"source_type"`
	SourceVersion  string `json:"source_version"`
	Status         string `json:"status"`
	CurrentPhase   string `json:"current_phase"`
	Checkpoint     int64  `json:"checkpoint"`
	Total          int64  `json:"total"`
	Processed      int64  `json:"processed"`
	SuccessCount   int64  `json:"success_count"`
	FailCount      int64  `json:"fail_count"`
	ErrorSummary   string `json:"error_summary"`
	FatalError     string `json:"fatal_error"`
	IdempotencyKey string `json:"idempotency_key"`
}

type RetryFailedResponse struct {
	Requeued bool   `json:"requeued"`
	Message  string `json:"message"`
}
