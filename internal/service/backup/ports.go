package service

import (
	"context"

	"github.com/gin-gonic/gin"
	commonService "github.com/lin-snow/ech0/internal/service/common"
)

type Service interface {
	ExportBackup(ctx *gin.Context, reqCtx context.Context) error
	CreateSnapshot(ctx context.Context) error
}

type CommonService = commonService.Service
