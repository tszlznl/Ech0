package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	snapshotTasks  sync.Map // map[string]commonModel.SnapshotTaskStatusResult
}

const backupS3UploadTimeout = 60 * time.Minute

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

	bs.tryUploadBackupToS3(backupFilePath, backupFileName)

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

func (bs *BackupService) CreateSnapshot(
	ctx context.Context,
) (*commonModel.SnapshotTaskCreateResult, error) {
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := bs.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return nil, err
	}
	if !user.IsAdmin {
		return nil, errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	taskID := uuid.NewString()
	startedAt := time.Now().UTC()
	bs.saveSnapshotTaskStatus(&commonModel.SnapshotTaskStatusResult{
		TaskID:    taskID,
		Status:    commonModel.SnapshotTaskStatusPending,
		StartedAt: startedAt,
		UpdatedAt: startedAt,
	})

	go bs.runSnapshotTask(taskID, userid)

	return &commonModel.SnapshotTaskCreateResult{
		TaskID: taskID,
		Status: commonModel.SnapshotTaskStatusPending,
	}, nil
}

func (bs *BackupService) GetSnapshotTaskStatus(
	ctx context.Context,
	taskID string,
) (*commonModel.SnapshotTaskStatusResult, error) {
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := bs.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return nil, err
	}
	if !user.IsAdmin {
		return nil, errors.New(commonModel.NO_PERMISSION_DENIED)
	}
	if taskID == "" {
		return nil, errors.New(commonModel.INVALID_PARAMS)
	}
	status, ok := bs.loadSnapshotTaskStatus(taskID)
	if !ok {
		return nil, errors.New(commonModel.INVALID_PARAMS)
	}
	return status, nil
}

func (bs *BackupService) runSnapshotTask(taskID, userID string) {
	startedAt := time.Now().UTC()
	if existing, ok := bs.loadSnapshotTaskStatus(taskID); ok {
		startedAt = existing.StartedAt
	}
	bs.saveSnapshotTaskStatus(&commonModel.SnapshotTaskStatusResult{
		TaskID:    taskID,
		Status:    commonModel.SnapshotTaskStatusRunning,
		StartedAt: startedAt,
		UpdatedAt: time.Now().UTC(),
	})

	taskCtx := viewer.WithContext(context.Background(), viewer.NewUserViewer(userID))
	backupFilePath, backupFileName, err := backup.ExecuteBackup()
	if err != nil {
		bs.failSnapshotTask(taskID, startedAt, err)
		return
	}

	// S3 上传失败不影响本地快照创建成功，失败仅记录日志。
	bs.tryUploadBackupToS3(backupFilePath, backupFileName)

	if err := bs.publisher.SystemBackup(
		taskCtx,
		contracts.SystemBackupEvent{Info: "System manual snapshot completed"},
	); err != nil {
		logUtil.GetLogger().Error("Failed to publish system backup completed event", zap.Error(err))
	}

	bs.saveSnapshotTaskStatus(&commonModel.SnapshotTaskStatusResult{
		TaskID:    taskID,
		Status:    commonModel.SnapshotTaskStatusSuccess,
		StartedAt: startedAt,
		UpdatedAt: time.Now().UTC(),
	})
}

func (bs *BackupService) failSnapshotTask(taskID string, startedAt time.Time, err error) {
	errMsg := commonModel.SNAPSHOT_UPLOAD_FAILED
	if err != nil {
		errMsg = err.Error()
	}
	bs.saveSnapshotTaskStatus(&commonModel.SnapshotTaskStatusResult{
		TaskID:    taskID,
		Status:    commonModel.SnapshotTaskStatusFailed,
		StartedAt: startedAt,
		UpdatedAt: time.Now().UTC(),
		Error:     errMsg,
	})
	logUtil.GetLogger().Warn("Snapshot task failed", zap.String("taskID", taskID), zap.Error(err))
}

func (bs *BackupService) loadSnapshotTaskStatus(
	taskID string,
) (*commonModel.SnapshotTaskStatusResult, bool) {
	if taskID == "" {
		return nil, false
	}
	value, ok := bs.snapshotTasks.Load(taskID)
	if !ok {
		return nil, false
	}
	status, ok := value.(commonModel.SnapshotTaskStatusResult)
	if !ok {
		return nil, false
	}
	cloned := status
	return &cloned, true
}

func (bs *BackupService) saveSnapshotTaskStatus(status *commonModel.SnapshotTaskStatusResult) {
	if status == nil || status.TaskID == "" {
		return
	}
	cloned := *status
	bs.snapshotTasks.Store(status.TaskID, cloned)
}

func (bs *BackupService) tryUploadBackupToS3(backupPath, fileName string) {
	if bs.storageManager == nil {
		return
	}
	selector := bs.storageManager.GetSelector()
	if selector == nil || !selector.ObjectEnabled() {
		return
	}

	uploadCtx, cancel := context.WithTimeout(context.Background(), backupS3UploadTimeout)
	defer cancel()

	cfg := bs.storageManager.GetStorageConfig(uploadCtx)
	if err := backup.UploadToS3(uploadCtx, backupPath, fileName, cfg); err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			logUtil.GetLogger().Warn(
				"Failed to upload backup to S3: upload timeout reached",
				zap.Duration("timeout", backupS3UploadTimeout),
				zap.Error(err),
			)
		case errors.Is(err, context.Canceled):
			logUtil.GetLogger().Warn(
				"Failed to upload backup to S3: upload context canceled",
				zap.Error(err),
			)
		default:
			logUtil.GetLogger().Warn("Failed to upload backup to S3", zap.Error(err))
		}
	}
}
