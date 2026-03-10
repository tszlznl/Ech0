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

func (h *MigrationHandler) CreateJob() gin.HandlerFunc {
	return response.Execute(func(ctx *gin.Context) response.Response {
		var req migrationModel.CreateMigrationJobRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			return response.Response{Msg: commonModel.INVALID_REQUEST_BODY, Err: err}
		}

		data, err := h.migrationService.CreateJob(ctx.Request.Context(), req)
		if err != nil {
			return response.Response{Msg: "", Err: err}
		}
		return response.Response{Msg: commonModel.SUCCESS_MESSAGE, Data: data}
	})
}

func (h *MigrationHandler) GetJob() gin.HandlerFunc {
	return response.Execute(func(ctx *gin.Context) response.Response {
		id := ctx.Param("id")
		data, err := h.migrationService.GetJob(ctx.Request.Context(), id)
		if err != nil {
			return response.Response{Msg: "", Err: err}
		}
		return response.Response{Msg: commonModel.SUCCESS_MESSAGE, Data: data}
	})
}

func (h *MigrationHandler) CancelJob() gin.HandlerFunc {
	return response.Execute(func(ctx *gin.Context) response.Response {
		id := ctx.Param("id")
		if err := h.migrationService.CancelJob(ctx.Request.Context(), id); err != nil {
			return response.Response{Msg: "", Err: err}
		}
		return response.Response{Msg: commonModel.SUCCESS_MESSAGE}
	})
}

func (h *MigrationHandler) RetryFailed() gin.HandlerFunc {
	return response.Execute(func(ctx *gin.Context) response.Response {
		id := ctx.Param("id")
		data, err := h.migrationService.RetryFailed(ctx.Request.Context(), id)
		if err != nil {
			return response.Response{Msg: "", Err: err}
		}
		return response.Response{Msg: commonModel.SUCCESS_MESSAGE, Data: data}
	})
}
