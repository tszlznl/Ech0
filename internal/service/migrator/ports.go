// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"mime/multipart"

	"github.com/lin-snow/ech0/internal/cache"
	"github.com/lin-snow/ech0/internal/kvstore"
	migrationModel "github.com/lin-snow/ech0/internal/model/migration"
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
	CommonService  = commonService.Service
	KVStore        = kvstore.Store
	StorageManager = *storage.Manager
	AppCache       = cache.ICache[string, any]
)
