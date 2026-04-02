package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/backup"
	contracts "github.com/lin-snow/ech0/internal/event/contracts"
	publisher "github.com/lin-snow/ech0/internal/event/publisher"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	"github.com/lin-snow/ech0/internal/storage"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"github.com/lin-snow/ech0/pkg/viewer"
	"go.uber.org/zap"
)

type BackupService struct {
	commonService  CommonService
	publisher      *publisher.Publisher
	storageManager *storage.Manager
}

func NewBackupService(
	commonService CommonService,
	publisher *publisher.Publisher,
	storageManager *storage.Manager,
) *BackupService {
	return &BackupService{
		commonService:  commonService,
		publisher:      publisher,
		storageManager: storageManager,
	}
}

func (bs *BackupService) ExportBackup(ctx *gin.Context, reqCtx context.Context) error {
	userid := viewer.MustFromContext(reqCtx).UserID()
	user, err := bs.commonService.CommonGetUserByUserId(reqCtx, userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	backupFilePath, backupFileName, err := backup.ExecuteBackup()
	if err != nil {
		return err
	}

	bs.tryUploadBackupToS3(reqCtx, backupFilePath, backupFileName)

	fileInfo, err := os.Stat(backupFilePath)
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("ech0-backup-%s.zip", time.Now().UTC().Format("2006-01-02-150405"))

	ctx.Writer.Header().Set("Content-Type", "application/zip")
	ctx.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	ctx.Writer.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	ctx.Writer.Header().Set("Accept-Ranges", "bytes")
	ctx.Writer.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	ctx.Writer.WriteHeader(200)
	ctx.File(backupFilePath)

	if err := bs.publisher.SystemExport(
		context.Background(),
		contracts.SystemExportEvent{
			Info: "System export completed",
			Size: fileInfo.Size(),
		},
	); err != nil {
		logUtil.GetLogger().Error("Failed to publish system export completed event", zap.Error(err))
	}

	return nil
}

func (bs *BackupService) CreateSnapshot(ctx context.Context) error {
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := bs.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	backupFilePath, backupFileName, err := backup.ExecuteBackup()
	if err != nil {
		return err
	}

	// S3 上传失败不影响本地快照创建成功，失败仅记录日志。
	bs.tryUploadBackupToS3(ctx, backupFilePath, backupFileName)

	if err := bs.publisher.SystemBackup(
		context.Background(),
		contracts.SystemBackupEvent{Info: "System manual snapshot completed"},
	); err != nil {
		logUtil.GetLogger().Error("Failed to publish system backup completed event", zap.Error(err))
	}

	return nil
}

func (bs *BackupService) tryUploadBackupToS3(ctx context.Context, backupPath, fileName string) {
	if bs.storageManager == nil {
		return
	}
	selector := bs.storageManager.GetSelector()
	if selector == nil || !selector.ObjectEnabled() {
		return
	}

	cfg := bs.storageManager.GetStorageConfig(ctx)
	s3FS, err := backup.BuildBackupS3FS(cfg)
	if err != nil {
		logUtil.GetLogger().Warn("Failed to build S3 FS for backup upload", zap.Error(err))
		return
	}

	if err := backup.UploadToS3(ctx, backupPath, fileName, s3FS); err != nil {
		logUtil.GetLogger().Warn("Failed to upload backup to S3", zap.Error(err))
	}
}
