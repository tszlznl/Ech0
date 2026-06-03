// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

// 导入来源类型(对应 importer/ 下的适配器)。
const (
	MigrationSourceEch0  = "ech0"
	MigrationSourceMemos = "memos"
)

// 导出目的地类型(对应 exporter/ 下的适配器):fs=本地目录,s3=对象存储。
const (
	ExportDestFS = "fs"
	ExportDestS3 = "s3"
)

// 导出阶段(异步导出作业用,与 MigrationPhase* 对称)。S3 上传移出关键路径后台跑,不再是作业阶段,
// 故只有 packing/completed 两段:本地产物一落盘即 completed,下载立即可用。
const (
	ExportPhasePacking   = "packing"
	ExportPhaseCompleted = "completed"
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
