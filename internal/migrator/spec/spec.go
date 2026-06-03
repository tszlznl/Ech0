// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package spec 定义 Migrator 引擎的对称契约:Importer(按来源把数据导入当前实例)与
// Exporter(按目的地把当前实例导出为 Snapshot)。二者镜像对称,各有一组 Request/Result。
package spec

import "context"

// ---- 导入端 ----

// ImportRequest 是导入的输入:来源载荷 + 进度回调。
type ImportRequest struct {
	SourcePayload  map[string]any
	UpdateProgress func(progress ImportProgress)
}

// ImportProgress 是导入的实时进度。
type ImportProgress struct {
	CurrentPhase string
	Processed    int64
	Total        int64
	SuccessCount int64
	FailCount    int64
	ErrorSummary string
}

// FailedItem 是单条导入失败记录(进 ImportResult.Report)。
type FailedItem struct {
	SourceID string `json:"source_id"`
	Reason   string `json:"reason"`
}

// ImportResult 是导入的结果汇总。
type ImportResult struct {
	Processed    int64
	Total        int64
	SuccessCount int64
	FailCount    int64
	ErrorSummary string
	JobID        string
	Report       map[string]any
}

// Importer 是导入端的来源适配器(ech0 / memos ...):把某个来源的数据载入当前实例。
type Importer interface {
	Import(ctx context.Context, req ImportRequest) (ImportResult, error)
}

// ---- 导出端 ----

// ExportRequest 是导出的输入(进度回调),与 ImportRequest 对称。
type ExportRequest struct {
	UpdateProgress func(progress ExportProgress)
}

// ExportProgress 是导出的实时进度,与 ImportProgress 对称。
type ExportProgress struct {
	CurrentPhase string
}

// ExportResult 是导出的产物描述:本地归档路径(供同步下载取回)、文件名、大小。
type ExportResult struct {
	ArtifactPath string
	FileName     string
	Size         int64
}

// Exporter 是导出端的目的地适配器(fs / s3 ...):把当前实例产出为一个 Snapshot 并落到目的地。
type Exporter interface {
	Export(ctx context.Context, req ExportRequest) (ExportResult, error)
}
