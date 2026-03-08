package service

import (
	"context"
	"mime/multipart"
	"time"

	"github.com/gin-gonic/gin"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	"github.com/lin-snow/ech0/internal/storage"
)

type Service interface {
	UploadFile(
		userID string,
		file *multipart.FileHeader,
		category storage.Category,
		storageType storage.StorageType,
	) (commonModel.FileDto, error)
	CreateExternalFile(userid string, dto commonModel.CreateExternalFileDto) (commonModel.FileDto, error)
	DeleteFile(userid string, dto commonModel.FileDeleteDto) error
	UploadAudioFile(userID string, file *multipart.FileHeader) (commonModel.FileDto, error)
	DeleteAudioFile(userid string) error
	GetCurrentAudioURL() string
	StreamCurrentAudio(ctx *gin.Context)
	GetFilePresignURL(userid string, dto *commonModel.GetPresignURLDto) (commonModel.PresignDto, error)
	CleanupOrphanFiles() error
	DeleteFileRecord(ctx context.Context, id string) error
	DeleteStoredFile(storageType string, key string) error
}

type CommonRepository interface {
	GetUserByUserId(ctx context.Context, id string) (userModel.User, error)
}

type KeyValueRepository interface {
	GetKeyValue(ctx context.Context, key string) (string, error)
	AddOrUpdateKeyValue(ctx context.Context, key, value string) error
	DeleteKeyValue(ctx context.Context, key string) error
}

type FileRepository interface {
	Create(ctx context.Context, file *fileModel.File) error
	GetByID(ctx context.Context, id string) (*fileModel.File, error)
	GetByKey(ctx context.Context, key string) (*fileModel.File, error)
	GetByRoute(ctx context.Context, storageType, provider, bucket, key string) (*fileModel.File, error)
	GetOrphanFiles(ctx context.Context, before time.Time) ([]fileModel.File, error)
	Delete(ctx context.Context, id string) error
	DeleteByRoute(ctx context.Context, storageType, provider, bucket, key string) error
}
