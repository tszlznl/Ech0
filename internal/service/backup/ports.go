package service

import (
	"mime/multipart"

	"github.com/gin-gonic/gin"
	commonService "github.com/lin-snow/ech0/internal/service/common"
)

type Service interface {
	Backup(userid string) error
	ExportBackup(ctx *gin.Context, userid string) error
	ImportBackup(ctx *gin.Context, userid string, file *multipart.FileHeader) error
}

type CommonService = commonService.Service
