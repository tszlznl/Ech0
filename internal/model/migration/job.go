package model

import "time"

const (
	MigrationSourceEch0V4 = "ech0_v4"
	MigrationSourceMemos  = "memos"
	MigrationSourceEch0V3 = "ech0_v3"
)

const (
	MigrationStatusIdle      = "idle"
	MigrationStatusPending   = "pending"
	MigrationStatusRunning   = "running"
	MigrationStatusSuccess   = "success"
	MigrationStatusFailed    = "failed"
	MigrationStatusCancelled = "cancelled"
)

const (
	MigrationPhaseExtracting   = "extracting"
	MigrationPhaseTransforming = "transforming"
	MigrationPhaseValidating   = "validating"
	MigrationPhaseLoading      = "loading"
	MigrationPhaseReporting    = "reporting"
	MigrationPhaseCompleted    = "completed"
)

type MigrationJob struct {
	ID           string     `gorm:"primaryKey;size:64" json:"id"`
	SourceType   string     `gorm:"type:varchar(64);index" json:"source_type"`
	Status       string     `gorm:"type:varchar(32);index" json:"status"`
	ErrorMessage string     `gorm:"type:text" json:"error_message"`
	StartedAt    *time.Time `json:"started_at"`
	FinishedAt   *time.Time `json:"finished_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
