package service

import (
	"context"
	"mime/multipart"

	migrationModel "github.com/lin-snow/ech0/internal/model/migration"
	keyvalueRepository "github.com/lin-snow/ech0/internal/repository/keyvalue"
	commonService "github.com/lin-snow/ech0/internal/service/common"
	"github.com/lin-snow/ech0/internal/storage"
)

type Service interface {
	UploadSourceZip(ctx context.Context, sourceType string, file *multipart.FileHeader) (migrationModel.UploadMigrationSourceZipResponse, error)
	StartGlobalMigration(ctx context.Context, req migrationModel.StartGlobalMigrationRequest) (migrationModel.GlobalMigrationStateDTO, error)
	GetGlobalMigrationStatus(ctx context.Context) (migrationModel.GlobalMigrationStateDTO, error)
	CancelGlobalMigration(ctx context.Context) (migrationModel.GlobalMigrationStateDTO, error)
	CleanupGlobalMigration(ctx context.Context) error
}

type (
	CommonService      = commonService.Service
	KeyValueRepository = *keyvalueRepository.KeyValueRepository
	StorageManager     = *storage.Manager
)
