// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

// MigrationPayload 是迁移作业的领域 payload，序列化进通用 Job 的 Payload 列：既是
// 提交时的输入，也是 MigrationRunner 干活时认得的结构。
type MigrationPayload struct {
	SourceType    string         `json:"source_type"`
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
	Version    int    `json:"version"`
	SourceType string `json:"source_type"`
	Status     string `json:"status"`
	// Phase 是细粒度阶段（extracting/loading/...），迁入 job 子系统后的净增字段；
	// 前端忽略未知字段，故向后兼容。
	Phase         string         `json:"phase,omitempty"`
	ErrorMessage  string         `json:"error_message"`
	SourcePayload map[string]any `json:"source_payload,omitempty"`
	StartedAt     *int64         `json:"started_at,omitempty"`
	UpdatedAt     *int64         `json:"updated_at,omitempty"`
	FinishedAt    *int64         `json:"finished_at,omitempty"`
}

// ExportPayload 是导出作业的领域 payload（序列化进通用 Job 的 Payload 列）。导出只有一种
// 原生格式（data/ 的 Snapshot），无需输入参数，故为空标记结构。
type ExportPayload struct{}

// ExportStateDTO 是导出作业对前端的状态契约，与 GlobalMigrationStateDTO 对称（status 复用
// MigrationStatus* 常量，含 idle 哨兵）。FileName/Size 在终态成功时由作业 Payload 补出。
type ExportStateDTO struct {
	Version      int    `json:"version"`
	Status       string `json:"status"`
	Phase        string `json:"phase,omitempty"`
	ErrorMessage string `json:"error_message"`
	FileName     string `json:"file_name,omitempty"`
	Size         int64  `json:"size,omitempty"`
	StartedAt    *int64 `json:"started_at,omitempty"`
	UpdatedAt    *int64 `json:"updated_at,omitempty"`
	FinishedAt   *int64 `json:"finished_at,omitempty"`
}
