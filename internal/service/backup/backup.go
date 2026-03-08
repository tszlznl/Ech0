package service

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/backup"
	contracts "github.com/lin-snow/ech0/internal/event/contracts"
	publisher "github.com/lin-snow/ech0/internal/event/publisher"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"go.uber.org/zap"
)

type BackupService struct {
	commonService CommonService
	publisher     *publisher.Publisher
}

func NewBackupService(
	commonService CommonService,
	publisher *publisher.Publisher,
) *BackupService {
	return &BackupService{
		commonService: commonService,
		publisher:     publisher,
	}
}

func (bs *BackupService) Backup(userid string) error {
	user, err := bs.commonService.CommonGetUserByUserId(context.Background(), userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	if _, _, err := backup.ExecuteBackup(); err != nil {
		return err
	}

	if err := bs.publisher.SystemBackup(
		context.Background(),
		contracts.SystemBackupEvent{Info: "System backup completed"},
	); err != nil {
		logUtil.GetLogger().Error("Failed to publish system backup completed event", zap.String("error", err.Error()))
	}

	return nil
}

func (bs *BackupService) ExportBackup(ctx *gin.Context, userid string) error {
	user, err := bs.commonService.CommonGetUserByUserId(context.Background(), userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	backupFilePath, _, err := backup.ExecuteBackup()
	if err != nil {
		return err
	}

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
		logUtil.GetLogger().Error("Failed to publish system export completed event", zap.String("error", err.Error()))
	}

	return nil
}

func (bs *BackupService) ImportBackup(
	ctx *gin.Context,
	userid string,
	file *multipart.FileHeader,
) error {
	user, err := bs.commonService.CommonGetUserByUserId(context.Background(), userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	if err := os.MkdirAll("./temp", 0o755); err != nil {
		return errors.New(commonModel.SNAPSHOT_UPLOAD_FAILED + ": " + err.Error())
	}

	timestamp := time.Now().UTC().Unix()
	tempFilePath := fmt.Sprintf("./temp/snapshot_%d.zip", timestamp)
	if err := ctx.SaveUploadedFile(file, tempFilePath); err != nil {
		return errors.New(commonModel.SNAPSHOT_UPLOAD_FAILED + ": " + err.Error())
	}

	if err := backup.ExcuteRestoreOnline(tempFilePath, timestamp); err != nil {
		return errors.New(commonModel.SNAPSHOT_RESTORE_FAILED + ": " + err.Error())
	}

	if err := bs.publisher.SystemRestore(
		context.Background(),
		contracts.SystemRestoreEvent{Info: "System restore completed"},
	); err != nil {
		logUtil.GetLogger().Error("Failed to publish system restore completed event", zap.String("error", err.Error()))
	}

	return nil
}
