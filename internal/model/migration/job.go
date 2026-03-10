package model

import "time"

const (
	MigrationSourceMemos  = "memos"
	MigrationSourceEch0V3 = "ech0_v3"
)

const (
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
	ID             string     `gorm:"primaryKey;size:64" json:"id"`
	SourceType     string     `gorm:"type:varchar(64);index:idx_migration_jobs_status_created,priority:3" json:"source_type"`
	SourceVersion  string     `gorm:"type:varchar(64)" json:"source_version"`
	Status         string     `gorm:"type:varchar(32);index:idx_migration_jobs_status_created,priority:1" json:"status"`
	CurrentPhase   string     `gorm:"type:varchar(32)" json:"current_phase"`
	SourcePayload  []byte     `gorm:"type:json" json:"source_payload"`
	IdempotencyKey string     `gorm:"type:varchar(255);index" json:"idempotency_key"`
	Checkpoint     int64      `gorm:"default:0" json:"checkpoint"`
	Total          int64      `gorm:"default:0" json:"total"`
	Processed      int64      `gorm:"default:0" json:"processed"`
	SuccessCount   int64      `gorm:"default:0" json:"success_count"`
	FailCount      int64      `gorm:"default:0" json:"fail_count"`
	ErrorSummary   string     `gorm:"type:text" json:"error_summary"`
	FatalError     string     `gorm:"type:text" json:"fatal_error"`
	FailedItems    []byte     `gorm:"type:json" json:"failed_items"`
	Report         []byte     `gorm:"type:json" json:"report"`
	CreatedBy      string     `gorm:"type:varchar(64);index:idx_migration_jobs_status_created,priority:2" json:"created_by"`
	StartedAt      *time.Time `json:"started_at"`
	FinishedAt     *time.Time `json:"finished_at"`
	LastHeartbeat  *time.Time `json:"last_heartbeat"`
	CreatedAt      time.Time  `gorm:"index:idx_migration_jobs_status_created,priority:4" json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
