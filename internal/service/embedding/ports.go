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

// Service 是 Embedding 索引与检索的对外接口。
type Service interface {
	IndexEcho(ctx context.Context, echo echoModel.Echo) error
	RemoveEcho(ctx context.Context, echoID string) error
	Backfill(ctx context.Context) (BackfillResult, error)
	Search(ctx context.Context, query string, k int) ([]model.SearchResult, error)
	Enabled(ctx context.Context) bool
	GetSetting(ctx context.Context) (settingModel.EmbeddingSetting, error)
	UpdateSetting(ctx context.Context, dto settingModel.EmbeddingSettingDto) error
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
	Search(ctx context.Context, vector []float32, k int) ([]model.SearchResult, error)
	ClearAll(ctx context.Context) error
	Count(ctx context.Context) (int64, error)
}

// KeyValueRepository 用于读写 Embedding 设置与索引状态（KeyValue 表）。
type KeyValueRepository interface {
	GetKeyValue(ctx context.Context, key string) (string, error)
	AddOrUpdateKeyValue(ctx context.Context, key string, value string) error
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
