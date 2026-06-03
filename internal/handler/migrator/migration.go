// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"github.com/gin-gonic/gin"
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

func (h *MigrationHandler) StartMigration() gin.HandlerFunc {
	return response.Execute(func(ctx *gin.Context) response.Response {
		var req migratorModel.StartGlobalMigrationRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			return response.Response{Msg: commonModel.INVALID_REQUEST_BODY, Err: err}
		}

		data, err := h.migrationService.StartGlobalMigration(ctx.Request.Context(), req)
		if err != nil {
			return response.Response{Msg: "", Err: err}
		}
		return response.Response{Msg: commonModel.SUCCESS_MESSAGE, Data: data}
	})
}

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

func (h *MigrationHandler) GetMigrationStatus() gin.HandlerFunc {
	return response.Execute(func(ctx *gin.Context) response.Response {
		data, err := h.migrationService.GetGlobalMigrationStatus(ctx.Request.Context())
		if err != nil {
			return response.Response{Msg: "", Err: err}
		}
		return response.Response{Msg: commonModel.SUCCESS_MESSAGE, Data: data}
	})
}

func (h *MigrationHandler) CancelMigration() gin.HandlerFunc {
	return response.Execute(func(ctx *gin.Context) response.Response {
		data, err := h.migrationService.CancelGlobalMigration(ctx.Request.Context())
		if err != nil {
			return response.Response{Msg: "", Err: err}
		}
		return response.Response{Msg: commonModel.SUCCESS_MESSAGE, Data: data}
	})
}

func (h *MigrationHandler) CleanupMigration() gin.HandlerFunc {
	return response.Execute(func(ctx *gin.Context) response.Response {
		if err := h.migrationService.CleanupGlobalMigration(ctx.Request.Context()); err != nil {
			return response.Response{Msg: "", Err: err}
		}
		return response.Response{Msg: commonModel.SUCCESS_MESSAGE}
	})
}

// StartExport 提交一次导出作业（手动快照异步出口，统一收敛到 Migrator 导出）。
func (h *MigrationHandler) StartExport() gin.HandlerFunc {
	return response.Execute(func(ctx *gin.Context) response.Response {
		data, err := h.migrationService.StartExport(ctx.Request.Context())
		if err != nil {
			return response.Response{Msg: "", Err: err}
		}
		return response.Response{Msg: commonModel.SUCCESS_MESSAGE, Data: data}
	})
}

// GetExportStatus 查询导出作业状态（查无作业行时为 idle 哨兵）。
func (h *MigrationHandler) GetExportStatus() gin.HandlerFunc {
	return response.Execute(func(ctx *gin.Context) response.Response {
		data, err := h.migrationService.GetExportStatus(ctx.Request.Context())
		if err != nil {
			return response.Response{Msg: "", Err: err}
		}
		return response.Response{Msg: commonModel.SUCCESS_MESSAGE, Data: data}
	})
}

// CancelExport 协作式取消在跑导出作业。
func (h *MigrationHandler) CancelExport() gin.HandlerFunc {
	return response.Execute(func(ctx *gin.Context) response.Response {
		data, err := h.migrationService.CancelExport(ctx.Request.Context())
		if err != nil {
			return response.Response{Msg: "", Err: err}
		}
		return response.Response{Msg: commonModel.SUCCESS_MESSAGE, Data: data}
	})
}

// DownloadExport 同步导出并触发浏览器下载。
//
//	@Summary		导出快照（下载）
//	@Description	同步产出快照并以附件形式下载
//	@Tags			数据迁移
//	@Accept			json
//	@Produce		application/octet-stream
//	@Success		200	{file}		file	"导出快照成功，返回文件下载"
//	@Failure		200	{string}	string	"导出快照失败"
//	@Router			/migration/export/download [get]
func (h *MigrationHandler) DownloadExport() gin.HandlerFunc {
	return response.Execute(func(ctx *gin.Context) response.Response {
		if err := h.migrationService.DownloadExport(ctx, ctx.Request.Context()); err != nil {
			return response.Response{Msg: "", Err: err}
		}
		return response.Response{Msg: commonModel.EXPORT_SNAPSHOT_SUCCESS}
	})
}
