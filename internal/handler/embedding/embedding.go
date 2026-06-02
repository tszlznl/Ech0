// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package handler 暴露 Embedding 向量索引的 HTTP 接口。
//
// Embedding 设置（get/update）归口到 setting 域；此处仅保留索引操作。重建索引改为
// 异步作业：触发即返回，前端按类型轮询 /reindex/status，可中途取消。
package handler

import (
	"encoding/json"
	"errors"

	"github.com/gin-gonic/gin"
	res "github.com/lin-snow/ech0/internal/handler/response"
	"github.com/lin-snow/ech0/internal/job"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
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

// reindexStatusResponse 是 reindex 作业的状态响应。payload 用 RawMessage 内嵌成对象
// （承载 BackfillResult: total/indexed/skipped/failed），避免被转义成字符串。
type reindexStatusResponse struct {
	Status     string          `json:"status"`
	Phase      string          `json:"phase,omitempty"`
	Error      string          `json:"error,omitempty"`
	Payload    json.RawMessage `json:"payload,omitempty"`
	StartedAt  *int64          `json:"started_at,omitempty"`
	FinishedAt *int64          `json:"finished_at,omitempty"`
}

func mapJobToReindexStatus(jb jobModel.Job) reindexStatusResponse {
	resp := reindexStatusResponse{
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

// Reindex 提交一次全量向量索引回填作业（管理员）；起即返回，不再阻塞。
func (embeddingHandler *EmbeddingHandler) Reindex() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		jb, err := embeddingHandler.jobManager.Submit(ctx.Request.Context(), jobModel.TypeReindex, nil)
		if err != nil {
			return res.Response{Err: err}
		}
		return res.Response{Data: mapJobToReindexStatus(jb), Msg: commonModel.SUCCESS_MESSAGE}
	})
}

// ReindexStatus 查询重建索引作业状态（前端按 type 轮询，无需 id）。
// 查无作业行时合成 idle，供前端判断「无进行中重建」。
func (embeddingHandler *EmbeddingHandler) ReindexStatus() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		jb, err := embeddingHandler.jobManager.Get(ctx.Request.Context(), jobModel.TypeReindex)
		if errors.Is(err, job.ErrNotFound) {
			return res.Response{Data: reindexStatusResponse{Status: reindexStatusIdle}, Msg: commonModel.SUCCESS_MESSAGE}
		}
		if err != nil {
			return res.Response{Err: err}
		}
		return res.Response{Data: mapJobToReindexStatus(jb), Msg: commonModel.SUCCESS_MESSAGE}
	})
}

// CancelReindex 取消进行中的重建索引作业；返回最新状态（前端轮询收敛到 cancelled）。
func (embeddingHandler *EmbeddingHandler) CancelReindex() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		_ = embeddingHandler.jobManager.Cancel(jobModel.TypeReindex)
		jb, err := embeddingHandler.jobManager.Get(ctx.Request.Context(), jobModel.TypeReindex)
		if errors.Is(err, job.ErrNotFound) {
			return res.Response{Data: reindexStatusResponse{Status: reindexStatusIdle}, Msg: commonModel.SUCCESS_MESSAGE}
		}
		if err != nil {
			return res.Response{Err: err}
		}
		return res.Response{Data: mapJobToReindexStatus(jb), Msg: commonModel.SUCCESS_MESSAGE}
	})
}
