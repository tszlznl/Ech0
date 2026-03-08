package service

import (
	"mime/multipart"

	"github.com/gin-gonic/gin"
	commonService "github.com/lin-snow/ech0/internal/service/common"
)

type Service interface {
	Backup(userid uint) error
	ExportBackup(ctx *gin.Context, userid uint) error
	ImportBackup(ctx *gin.Context, userid uint, file *multipart.FileHeader) error
}

type CommonService = commonService.Service
