package handler

import (
	"github.com/gin-gonic/gin"
	response "github.com/lin-snow/ech0/internal/handler/response"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	migrationModel "github.com/lin-snow/ech0/internal/model/migration"
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
		var req migrationModel.StartGlobalMigrationRequest
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
