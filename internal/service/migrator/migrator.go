package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	migrationModel "github.com/lin-snow/ech0/internal/model/migration"
	migrationRepository "github.com/lin-snow/ech0/internal/repository/migration"
	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
	"github.com/lin-snow/ech0/pkg/viewer"
)

type MigratorService struct {
	commonService       CommonService
	migrationRepository MigrationRepository
}

func NewMigratorService(
	commonService CommonService,
	migrationRepository MigrationRepository,
) *MigratorService {
	return &MigratorService{
		commonService:       commonService,
		migrationRepository: migrationRepository,
	}
}

func (s *MigratorService) CreateJob(
	ctx context.Context,
	req migrationModel.CreateMigrationJobRequest,
) (migrationModel.MigrationJobDTO, error) {
	if err := validateCreateRequest(req); err != nil {
		return migrationModel.MigrationJobDTO{}, err
	}

	userID := viewer.MustFromContext(ctx).UserID()
	user, err := s.commonService.CommonGetUserByUserId(ctx, userID)
	if err != nil {
		return migrationModel.MigrationJobDTO{}, err
	}
	if !user.IsAdmin {
		return migrationModel.MigrationJobDTO{}, errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	payload, err := json.Marshal(req.SourcePayload)
	if err != nil {
		return migrationModel.MigrationJobDTO{}, fmt.Errorf("marshal source payload: %w", err)
	}

	jobID := uuidUtil.MustNewV7()
	idempotencyKey := fmt.Sprintf("%s:%s:%s", req.SourceType, req.SourceVersion, jobID)
	job := &migrationModel.MigrationJob{
		ID:             jobID,
		SourceType:     req.SourceType,
		SourceVersion:  req.SourceVersion,
		Status:         migrationModel.MigrationStatusPending,
		CurrentPhase:   migrationModel.MigrationPhaseExtracting,
		SourcePayload:  payload,
		IdempotencyKey: idempotencyKey,
		CreatedBy:      userID,
	}

	if err := s.migrationRepository.CreateJob(ctx, job); err != nil {
		return migrationModel.MigrationJobDTO{}, err
	}

	return toDTO(*job), nil
}

func (s *MigratorService) GetJob(ctx context.Context, id string) (migrationModel.MigrationJobDTO, error) {
	job, err := s.GetJobModel(ctx, id)
	if err != nil {
		return migrationModel.MigrationJobDTO{}, err
	}
	return toDTO(job), nil
}

func (s *MigratorService) GetJobModel(ctx context.Context, id string) (migrationModel.MigrationJob, error) {
	job, err := s.migrationRepository.GetJobByID(ctx, id)
	if err != nil {
		if errors.Is(err, migrationRepository.ErrMigrationJobNotFound) {
			return migrationModel.MigrationJob{}, errors.New(commonModel.MIGRATION_JOB_NOT_FOUND)
		}
		return migrationModel.MigrationJob{}, err
	}
	return job, nil
}

func (s *MigratorService) CancelJob(ctx context.Context, id string) error {
	userID := viewer.MustFromContext(ctx).UserID()
	user, err := s.commonService.CommonGetUserByUserId(ctx, userID)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	if err := s.migrationRepository.MarkCancelled(ctx, id); err != nil {
		if errors.Is(err, migrationRepository.ErrMigrationJobNotFound) {
			return errors.New(commonModel.MIGRATION_JOB_NOT_FOUND)
		}
		return err
	}
	return nil
}

func (s *MigratorService) RetryFailed(
	ctx context.Context,
	id string,
) (migrationModel.RetryFailedResponse, error) {
	userID := viewer.MustFromContext(ctx).UserID()
	user, err := s.commonService.CommonGetUserByUserId(ctx, userID)
	if err != nil {
		return migrationModel.RetryFailedResponse{}, err
	}
	if !user.IsAdmin {
		return migrationModel.RetryFailedResponse{}, errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	job, err := s.migrationRepository.GetJobByID(ctx, id)
	if err != nil {
		if errors.Is(err, migrationRepository.ErrMigrationJobNotFound) {
			return migrationModel.RetryFailedResponse{}, errors.New(commonModel.MIGRATION_JOB_NOT_FOUND)
		}
		return migrationModel.RetryFailedResponse{}, err
	}
	if job.Status != migrationModel.MigrationStatusFailed && job.Status != migrationModel.MigrationStatusCancelled {
		return migrationModel.RetryFailedResponse{}, errors.New(commonModel.INVALID_REQUEST_BODY)
	}

	now := time.Now().UTC()
	job.Status = migrationModel.MigrationStatusPending
	job.CurrentPhase = migrationModel.MigrationPhaseExtracting
	job.Checkpoint = 0
	job.Total = 0
	job.Processed = 0
	job.SuccessCount = 0
	job.FailCount = 0
	job.ErrorSummary = ""
	job.FatalError = ""
	job.Report = nil
	job.StartedAt = nil
	job.FinishedAt = nil
	job.LastHeartbeat = &now
	job.IdempotencyKey = fmt.Sprintf("%s:retry:%d", job.IdempotencyKey, now.Unix())

	if err := s.migrationRepository.UpdateJob(ctx, &job); err != nil {
		return migrationModel.RetryFailedResponse{}, err
	}

	return migrationModel.RetryFailedResponse{
		Requeued: true,
		Message:  "重试任务已加入队列",
	}, nil
}

func (s *MigratorService) ClaimNextPendingJob(ctx context.Context) (migrationModel.MigrationJob, error) {
	return s.migrationRepository.ClaimNextPendingJob(ctx)
}

func (s *MigratorService) UpdateJob(ctx context.Context, job *migrationModel.MigrationJob) error {
	return s.migrationRepository.UpdateJob(ctx, job)
}

func validateCreateRequest(req migrationModel.CreateMigrationJobRequest) error {
	sourceType := strings.TrimSpace(req.SourceType)
	switch sourceType {
	case migrationModel.MigrationSourceMemos, migrationModel.MigrationSourceEch0V3:
		return nil
	default:
		return errors.New(commonModel.INVALID_REQUEST_BODY)
	}
}

func toDTO(job migrationModel.MigrationJob) migrationModel.MigrationJobDTO {
	return migrationModel.MigrationJobDTO{
		ID:             job.ID,
		SourceType:     job.SourceType,
		SourceVersion:  job.SourceVersion,
		Status:         job.Status,
		CurrentPhase:   job.CurrentPhase,
		Checkpoint:     job.Checkpoint,
		Total:          job.Total,
		Processed:      job.Processed,
		SuccessCount:   job.SuccessCount,
		FailCount:      job.FailCount,
		ErrorSummary:   job.ErrorSummary,
		FatalError:     job.FatalError,
		IdempotencyKey: job.IdempotencyKey,
	}
}
