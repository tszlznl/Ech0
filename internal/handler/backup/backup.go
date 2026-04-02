package handler

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	res "github.com/lin-snow/ech0/internal/handler/response"
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
//	@Success		200	{file}		file	"导出备份成功，返回文件下载"
//	@Failure		200	{string}	string	"导出备份失败"
//	@Router			/backup/export [get]
func (backupHandler *BackupHandler) ExportBackup() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		token := ctx.Query("token")
		if token == "" {
			return res.Response{
				Msg: commonModel.INVALID_REQUEST_BODY,
			}
		}

		token = strings.Trim(token, `"`) // 去掉可能的双引号

		// 使用 JWT Util进行处理
		claims, err := jwtUtil.ParseToken(token)
		if err != nil {
			return res.Response{
				Msg: commonModel.TOKEN_NOT_VALID,
				Err: err,
			}
		}

		// 从 Claims中提取 UserID，注入 viewer 上下文供 service 鉴权。
		reqCtx := viewer.WithContext(context.Background(), viewer.NewUserViewer(claims.Userid))
		if err := backupHandler.backupService.ExportBackup(ctx, reqCtx); err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Msg: commonModel.EXPORT_BACKUP_SUCCESS,
		}
	})
}

// CreateSnapshot 手动创建快照
//
//	@Summary		手动创建快照
//	@Description	仅在服务端创建本地快照；若配置了 S3 则尝试上传（失败静默）
//	@Tags			系统备份
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	res.Response	"创建快照成功"
//	@Failure		200	{object}	res.Response	"创建快照失败"
//	@Router			/backup/snapshot [post]
func (backupHandler *BackupHandler) CreateSnapshot() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		task, err := backupHandler.backupService.CreateSnapshot(ctx.Request.Context())
		if err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Data: task,
			Msg:  commonModel.CREATE_SNAPSHOT_SUCCESS,
		}
	})
}

// GetSnapshotStatus 查询快照任务状态
//
//	@Summary		查询快照任务状态
//	@Description	根据 taskId 查询创建快照任务的执行状态
//	@Tags			系统备份
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	res.Response	"查询快照状态成功"
//	@Failure		200	{object}	res.Response	"查询快照状态失败"
//	@Router			/backup/snapshot/{taskId} [get]
func (backupHandler *BackupHandler) GetSnapshotStatus() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		taskID := ctx.Param("taskId")
		result, err := backupHandler.backupService.GetSnapshotTaskStatus(ctx.Request.Context(), taskID)
		if err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Data: result,
			Msg:  commonModel.SUCCESS_MESSAGE,
		}
	})
}
