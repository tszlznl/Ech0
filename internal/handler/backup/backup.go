package handler

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	response "github.com/lin-snow/ech0/internal/handler/response"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	service "github.com/lin-snow/ech0/internal/service/backup"
	jwtUtil "github.com/lin-snow/ech0/internal/util/jwt"
	"github.com/lin-snow/ech0/pkg/viewer"
)

type BackupHandler struct {
	backupService service.Service
}

// NewBackupHandler BackupHandler 的构造函数
func NewBackupHandler(backupService service.Service) *BackupHandler {
	return &BackupHandler{
		backupService: backupService,
	}
}

// ExportBackup 导出数据备份
//
//	@Summary		导出数据备份
//	@Description	用户导出备份文件，成功后触发文件下载
//	@Tags			系统备份
//	@Accept			json
//	@Produce		application/octet-stream
//	@Success		200	{object}	response.Response	"导出备份成功，返回文件下载"
//	@Failure		200	{object}	response.Response	"导出备份失败"
//	@Router			/backup/export [get]
func (backupHandler *BackupHandler) ExportBackup() gin.HandlerFunc {
	return response.Execute(func(ctx *gin.Context) response.Response {
		token := ctx.Query("token")
		if token == "" {
			return response.Response{
				Msg: commonModel.INVALID_REQUEST_BODY,
			}
		}

		token = strings.Trim(token, `"`) // 去掉可能的双引号

		// 使用 JWT Util进行处理
		claims, err := jwtUtil.ParseToken(token)
		if err != nil {
			return response.Response{
				Msg: commonModel.TOKEN_NOT_VALID,
				Err: err,
			}
		}

		// 从 Claims中提取 UserID，注入 viewer 上下文供 service 鉴权。
		reqCtx := viewer.WithContext(context.Background(), viewer.NewUserViewer(claims.Userid))
		if err := backupHandler.backupService.ExportBackup(ctx, reqCtx); err != nil {
			return response.Response{
				Msg: "",
				Err: err,
			}
		}

		return response.Response{
			Msg: commonModel.EXPORT_BACKUP_SUCCESS,
		}
	})
}
