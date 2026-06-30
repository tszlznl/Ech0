// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package service 实现 Embedding 的索引与检索逻辑。
package service

import (
	"context"

	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	model "github.com/lin-snow/ech0/internal/model/embedding"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
)

// Embedder 是文本向量化的可注入依赖（默认实现 internal/embedding.Client）。
// 抽出接口是为了让 IndexEcho/Search/Backfill 的"取向量"步骤在测试中可用替身，
// 不必真发 /v1/embeddings 请求，从而覆盖 seam 之后的 upsert / 作者收口 / 分页逻辑。
type Embedder interface {
	Embed(ctx context.Context, setting settingModel.EmbeddingSetting, inputs []string) ([][]float32, error)
	EmbedOne(ctx context.Context, setting settingModel.EmbeddingSetting, input string) ([]float32, error)
}

// Service 是 Embedding 索引与检索的对外接口。
//
// Embedding 设置的读写已移交 setting 域（settingService），故此接口不再暴露
// GetSetting/UpdateSetting；本服务内部仍按需读取该设置用于索引与检索。
type Service interface {
	IndexEcho(ctx context.Context, echo echoModel.Echo) error
	RemoveEcho(ctx context.Context, echoID string) error
	// Backfill 全量回填历史 Echo 的向量。onProgress 非 nil 时每页结束回调累计计数
	// （供异步 job 上报实时进度）；长循环尊重 ctx 取消（reindex job 可中断）。
	Backfill(ctx context.Context, onProgress func(BackfillResult)) (BackfillResult, error)
	// Search 做语义检索。authorUsername 非空时把命中收口到该作者发布的 Echo
	// （Copilot Chat 用它隔离多用户实例下的他人 Echo）；空串表示不限定作者。
	Search(ctx context.Context, query string, k int, authorUsername string) ([]model.SearchResult, error)
	Enabled(ctx context.Context) bool
}

// Indexer 是事件订阅者使用的最小接口（仅增量索引）。
type Indexer interface {
	IndexEcho(ctx context.Context, echo echoModel.Echo) error
	RemoveEcho(ctx context.Context, echoID string) error
}

// Repository 是向量存储（vec0 虚表 + 元数据表）的抽象。
type Repository interface {
	EnsureVecTable(ctx context.Context, dim int) error
	DropVecTable(ctx context.Context) error
	Upsert(ctx context.Context, meta *model.EchoEmbedding, vector []float32) error
	Delete(ctx context.Context, echoID string) error
	GetMeta(ctx context.Context, echoID string) (*model.EchoEmbedding, bool, error)
	// Search 做向量 KNN 检索。authorUsername 非空时把命中收口到该作者
	// （over-fetch 后按 username 过滤，仍返回最多 k 条）；空串表示不限定作者。
	Search(ctx context.Context, vector []float32, k int, authorUsername string) ([]model.SearchResult, error)
	ClearAll(ctx context.Context) error
	Count(ctx context.Context) (int64, error)
}

// EchoReader 用于回填时分页读取全部 Echo。
type EchoReader interface {
	GetEchosByPage(page, pageSize int, search string, showPrivate bool) ([]echoModel.Echo, int64)
}

// BackfillResult 是回填结果统计。
type BackfillResult struct {
	Total   int `json:"total"`
	Indexed int `json:"indexed"`
	Skipped int `json:"skipped"`
	Failed  int `json:"failed"`
}
