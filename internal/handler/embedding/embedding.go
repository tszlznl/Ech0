// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package handler 暴露 Embedding 向量索引的 HTTP 接口（Huma type-first）。
//
// Embedding 设置（get/update）归口到 setting 域；此处仅保留索引操作。重建索引为异步作业：
// 触发即返回，前端按类型轮询 /embedding/reindex/status，可中途取消。
package handler

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/lin-snow/ech0/internal/handler/humares"
	"github.com/lin-snow/ech0/internal/job"
	jobModel "github.com/lin-snow/ech0/internal/model/job"
)

// reindexStatusIdle 是「从未运行 / 已无作业行」时合成的哨兵状态，对应 job.ErrNotFound。
const reindexStatusIdle = "idle"

type EmbeddingHandler struct {
	jobManager *job.Manager
}

func NewEmbeddingHandler(jobManager *job.Manager) *EmbeddingHandler {
	return &EmbeddingHandler{
		jobManager: jobManager,
	}
}

// ReindexStatusResponse 是 reindex 作业的状态响应。payload 用 RawMessage 内嵌成对象
// （承载 BackfillResult: total/indexed/skipped/failed），避免被转义成字符串。
type ReindexStatusResponse struct {
	Status     string          `json:"status" doc:"作业状态：idle/pending/running/succeeded/failed/cancelled" example:"running"`
	Phase      string          `json:"phase,omitempty" doc:"当前阶段"`
	Error      string          `json:"error,omitempty" doc:"失败原因（status=failed 时）"`
	Payload    json.RawMessage `json:"payload,omitempty" doc:"回填结果 BackfillResult: total/indexed/skipped/failed"`
	StartedAt  *int64          `json:"started_at,omitempty" doc:"开始时间（Unix 秒）"`
	FinishedAt *int64          `json:"finished_at,omitempty" doc:"结束时间（Unix 秒）"`
}

// 空入参类型：这些操作不带 path/query/body 参数，但 Huma 要求每个 operation 有输入结构。
type (
	ReindexInput       struct{}
	ReindexStatusInput struct{}
	CancelReindexInput struct{}
)

func mapJobToReindexStatus(jb jobModel.Job) ReindexStatusResponse {
	resp := ReindexStatusResponse{
		Status:     string(jb.Status),
		Phase:      jb.Phase,
		Error:      jb.Error,
		StartedAt:  jb.StartedAt,
		FinishedAt: jb.FinishedAt,
	}
	if jb.Payload != "" {
		resp.Payload = json.RawMessage(jb.Payload)
	}
	return resp
}

// Reindex 提交一次全量向量索引回填作业（管理员）；起即返回，不阻塞。
func (embeddingHandler *EmbeddingHandler) Reindex(ctx context.Context, _ *ReindexInput) (*humares.Envelope[ReindexStatusResponse], error) {
	jb, err := embeddingHandler.jobManager.Submit(ctx, jobModel.TypeReindex, nil)
	if err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK(ctx, mapJobToReindexStatus(jb)), nil
}

// ReindexStatus 查询重建索引作业状态（前端按 type 轮询，无需 id）。
// 查无作业行时合成 idle，供前端判断「无进行中重建」。
func (embeddingHandler *EmbeddingHandler) ReindexStatus(ctx context.Context, _ *ReindexStatusInput) (*humares.Envelope[ReindexStatusResponse], error) {
	jb, err := embeddingHandler.jobManager.Get(ctx, jobModel.TypeReindex)
	if errors.Is(err, job.ErrNotFound) {
		return humares.OK(ctx, ReindexStatusResponse{Status: reindexStatusIdle}), nil
	}
	if err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK(ctx, mapJobToReindexStatus(jb)), nil
}

// CancelReindex 取消进行中的重建索引作业；返回最新状态（前端轮询收敛到 cancelled）。
func (embeddingHandler *EmbeddingHandler) CancelReindex(ctx context.Context, _ *CancelReindexInput) (*humares.Envelope[ReindexStatusResponse], error) {
	_ = embeddingHandler.jobManager.Cancel(jobModel.TypeReindex)
	jb, err := embeddingHandler.jobManager.Get(ctx, jobModel.TypeReindex)
	if errors.Is(err, job.ErrNotFound) {
		return humares.OK(ctx, ReindexStatusResponse{Status: reindexStatusIdle}), nil
	}
	if err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK(ctx, mapJobToReindexStatus(jb)), nil
}
