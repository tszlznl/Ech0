package model

import "time"

type CreateMigrationJobRequest struct {
	SourceType    string         `json:"source_type" binding:"required"`
	SourcePayload map[string]any `json:"source_payload"`
}

type StartGlobalMigrationRequest struct {
	SourceType    string         `json:"source_type" binding:"required"`
	SourcePayload map[string]any `json:"source_payload"`
}

type UploadMigrationSourceZipResponse struct {
	SourceType    string         `json:"source_type"`
	TmpDir        string         `json:"tmp_dir"`
	SourcePayload map[string]any `json:"source_payload"`
}

type GlobalMigrationStateDTO struct {
	Version       int            `json:"version"`
	SourceType    string         `json:"source_type"`
	Status        string         `json:"status"`
	ErrorMessage  string         `json:"error_message"`
	SourcePayload map[string]any `json:"source_payload,omitempty"`
	StartedAt     *time.Time     `json:"started_at,omitempty"`
	UpdatedAt     *time.Time     `json:"updated_at,omitempty"`
	FinishedAt    *time.Time     `json:"finished_at,omitempty"`
}
