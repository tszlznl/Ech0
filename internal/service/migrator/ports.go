package service

import (
	"context"

	migrationModel "github.com/lin-snow/ech0/internal/model/migration"
	migrationRepository "github.com/lin-snow/ech0/internal/repository/migration"
	commonService "github.com/lin-snow/ech0/internal/service/common"
)

type Service interface {
	CreateJob(ctx context.Context, req migrationModel.CreateMigrationJobRequest) (migrationModel.MigrationJobDTO, error)
	GetJob(ctx context.Context, id string) (migrationModel.MigrationJobDTO, error)
	GetJobModel(ctx context.Context, id string) (migrationModel.MigrationJob, error)
	CancelJob(ctx context.Context, id string) error
	RetryFailed(ctx context.Context, id string) (migrationModel.RetryFailedResponse, error)

	ClaimNextPendingJob(ctx context.Context) (migrationModel.MigrationJob, error)
	UpdateJob(ctx context.Context, job *migrationModel.MigrationJob) error
}

type CommonService = commonService.Service
type MigrationRepository = *migrationRepository.MigrationRepository
