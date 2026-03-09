package service

import (
	"context"
	"mime/multipart"

	"github.com/gin-gonic/gin"
	commonService "github.com/lin-snow/ech0/internal/service/common"
)

type Service interface {
	Backup(ctx context.Context) error
	ExportBackup(ctx *gin.Context, reqCtx context.Context) error
	ImportBackup(ctx *gin.Context, reqCtx context.Context, file *multipart.FileHeader) error
}

type CommonService = commonService.Service
