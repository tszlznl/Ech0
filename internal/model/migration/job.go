// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

const (
	MigrationSourceEch0V4 = "ech0_v4"
	MigrationSourceMemos  = "memos"
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
