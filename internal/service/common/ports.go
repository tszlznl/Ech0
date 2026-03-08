package service

import (
	"context"
	"mime/multipart"
	"time"

	"github.com/gin-gonic/gin"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	storageDomain "github.com/lin-snow/ech0/internal/storage"
)

type Service interface {
	CommonGetUserByUserId(ctx context.Context, userId string) (userModel.User, error)
	UploadFile(userId string, file *multipart.FileHeader, category storageDomain.Category) (commonModel.FileDto, error)
	DeleteFile(userid string, dto commonModel.FileDeleteDto) error
	GetSysAdmin() (userModel.User, error)
	GetStatus() (commonModel.Status, error)
	GetHeatMap(timezone string) ([]commonModel.Heatmap, error)
	GenerateRSS(ctx *gin.Context) (string, error)
	UploadMusic(userId string, file *multipart.FileHeader) (string, error)
	DeleteMusic(userid string) error
	GetPlayMusicUrl() string
	PlayMusic(ctx *gin.Context)
	GetFilePresignURL(userid string, s3Dto *commonModel.GetPresignURLDto, method string) (commonModel.PresignDto, error)
	CleanupOrphanFiles() error
	GetWebsiteTitle(websiteURL string) (string, error)
	DeleteFileRecord(ctx context.Context, id string) error
	DeleteStoredFile(key string) error
}

type CommonRepository interface {
	GetUserByUserId(ctx context.Context, id string) (userModel.User, error)
	GetSysAdmin(ctx context.Context) (userModel.User, error)
	GetAllUsers(ctx context.Context) ([]userModel.User, error)
	GetAllEchos(ctx context.Context, showPrivate bool) ([]echoModel.Echo, error)
	GetHeatMap(ctx context.Context, startTime, endTime time.Time) ([]time.Time, error)
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
