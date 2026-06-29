// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package handler 暴露数据迁移（导入/导出快照）的 HTTP 接口。
//
// 控制面 JSON 端点（start/status/cancel/cleanup/export）走 Huma type-first；
// multipart 上传源 zip 与二进制快照下载仍走裸 gin（见本文件下方 + setupMigrationRoutes）。
package handler

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/handler/humares"
	response "github.com/lin-snow/ech0/internal/handler/response"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	migratorModel "github.com/lin-snow/ech0/internal/model/migrator"
	service "github.com/lin-snow/ech0/internal/service/migrator"
)

type MigrationHandler struct {
	migrationService service.Service
}

func NewMigrationHandler(migrationService service.Service) *MigrationHandler {
	return &MigrationHandler{
		migrationService: migrationService,
	}
}

type (
	StartMigrationInput struct {
		Body migratorModel.StartGlobalMigrationRequest
	}
	GetMigrationStatusInput struct{}
	CancelMigrationInput    struct{}
	CleanupMigrationInput   struct{}
	StartExportInput        struct{}
	GetExportStatusInput    struct{}
	CancelExportInput       struct{}
)

// StartMigration 启动一次全局数据迁移（admin:settings）。
func (h *MigrationHandler) StartMigration(ctx context.Context, in *StartMigrationInput) (*humares.Envelope[migratorModel.GlobalMigrationStateDTO], error) {
	data, err := h.migrationService.StartGlobalMigration(ctx, in.Body)
	if err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK(ctx, data), nil
}

// GetMigrationStatus 查询全局迁移状态（admin:settings）。
func (h *MigrationHandler) GetMigrationStatus(ctx context.Context, _ *GetMigrationStatusInput) (*humares.Envelope[migratorModel.GlobalMigrationStateDTO], error) {
	data, err := h.migrationService.GetGlobalMigrationStatus(ctx)
	if err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK(ctx, data), nil
}

// CancelMigration 取消进行中的全局迁移（admin:settings）。
func (h *MigrationHandler) CancelMigration(ctx context.Context, _ *CancelMigrationInput) (*humares.Envelope[migratorModel.GlobalMigrationStateDTO], error) {
	data, err := h.migrationService.CancelGlobalMigration(ctx)
	if err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK(ctx, data), nil
}

// CleanupMigration 清理迁移中间产物（admin:settings）。
func (h *MigrationHandler) CleanupMigration(ctx context.Context, _ *CleanupMigrationInput) (*humares.Envelope[any], error) {
	if err := h.migrationService.CleanupGlobalMigration(ctx); err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK[any](ctx, nil), nil
}

// StartExport 提交一次导出作业（手动快照异步出口，admin:settings）。
func (h *MigrationHandler) StartExport(ctx context.Context, _ *StartExportInput) (*humares.Envelope[migratorModel.ExportStateDTO], error) {
	data, err := h.migrationService.StartExport(ctx)
	if err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK(ctx, data), nil
}

// GetExportStatus 查询导出作业状态（admin:settings）。
func (h *MigrationHandler) GetExportStatus(ctx context.Context, _ *GetExportStatusInput) (*humares.Envelope[migratorModel.ExportStateDTO], error) {
	data, err := h.migrationService.GetExportStatus(ctx)
	if err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK(ctx, data), nil
}

// CancelExport 协作式取消在跑导出作业（admin:settings）。
func (h *MigrationHandler) CancelExport(ctx context.Context, _ *CancelExportInput) (*humares.Envelope[migratorModel.ExportStateDTO], error) {
	data, err := h.migrationService.CancelExport(ctx)
	if err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK(ctx, data), nil
}

// --- 以下为非 JSON 端点，仍走裸 gin（multipart 上传 / 二进制快照下载） ---

func (h *MigrationHandler) UploadSourceZip() gin.HandlerFunc {
	return response.Execute(func(ctx *gin.Context) response.Response {
		sourceType := ctx.PostForm("source_type")
		if sourceType == "" {
			return response.Response{Msg: commonModel.INVALID_REQUEST_BODY}
		}
		file, err := ctx.FormFile("file")
		if err != nil {
			return response.Response{Msg: commonModel.INVALID_REQUEST_BODY, Err: err}
		}

		data, err := h.migrationService.UploadSourceZip(ctx.Request.Context(), sourceType, file)
		if err != nil {
			return response.Response{Msg: "", Err: err}
		}
		return response.Response{Msg: commonModel.SUCCESS_MESSAGE, Data: data}
	})
}

// DownloadExport 同步导出并触发浏览器下载（二进制 octet-stream）。
func (h *MigrationHandler) DownloadExport() gin.HandlerFunc {
	return response.Execute(func(ctx *gin.Context) response.Response {
		if err := h.migrationService.DownloadExport(ctx, ctx.Request.Context()); err != nil {
			return response.Response{Msg: "", Err: err}
		}
		return response.Response{Msg: commonModel.EXPORT_SNAPSHOT_SUCCESS}
	})
}
