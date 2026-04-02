package service

import (
	"context"

	"github.com/gin-gonic/gin"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	commonService "github.com/lin-snow/ech0/internal/service/common"
)

type Service interface {
	ExportBackup(ctx *gin.Context, reqCtx context.Context) error
	CreateSnapshot(ctx context.Context) (*commonModel.SnapshotTaskCreateResult, error)
	GetSnapshotTaskStatus(
		ctx context.Context,
		taskID string,
	) (*commonModel.SnapshotTaskStatusResult, error)
}

type CommonService = commonService.Service
