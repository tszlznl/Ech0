package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	coreMigrator "github.com/lin-snow/ech0/internal/migrator"
	"github.com/lin-snow/ech0/internal/migrator/spec"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/lin-snow/ech0/internal/backup"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	migrationModel "github.com/lin-snow/ech0/internal/model/migration"
	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
	"github.com/lin-snow/ech0/pkg/viewer"
	"gorm.io/gorm"
)

const migrationTmpRelativeDir = "files/tmp"

type MigratorService struct {
	commonService      CommonService
	keyValueRepository KeyValueRepository

	activeMu     sync.Mutex
	activeCancel context.CancelFunc
}

func NewMigratorService(
	commonService CommonService,
	keyValueRepository KeyValueRepository,
) *MigratorService {
	return &MigratorService{
		commonService:      commonService,
		keyValueRepository: keyValueRepository,
	}
}

func (s *MigratorService) UploadSourceZip(
	ctx context.Context,
	sourceType string,
	file *multipart.FileHeader,
) (migrationModel.UploadMigrationSourceZipResponse, error) {
	if err := validateSourceType(sourceType); err != nil {
		return migrationModel.UploadMigrationSourceZipResponse{}, err
	}
	if file == nil {
		return migrationModel.UploadMigrationSourceZipResponse{}, errors.New(commonModel.INVALID_REQUEST_BODY)
	}

	userID := viewer.MustFromContext(ctx).UserID()
	user, err := s.commonService.CommonGetUserByUserId(ctx, userID)
	if err != nil {
		return migrationModel.UploadMigrationSourceZipResponse{}, err
	}
	if !user.IsAdmin {
		return migrationModel.UploadMigrationSourceZipResponse{}, errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	if !strings.HasSuffix(strings.ToLower(file.Filename), ".zip") {
		return migrationModel.UploadMigrationSourceZipResponse{}, errors.New(commonModel.INVALID_REQUEST_BODY)
	}

	baseTmpDir := filepath.Join("data", migrationTmpRelativeDir)
	if err := os.MkdirAll(baseTmpDir, 0o755); err != nil {
		return migrationModel.UploadMigrationSourceZipResponse{}, fmt.Errorf("create migration tmp dir: %w", err)
	}

	uploadID := uuidUtil.MustNewV7()
	folderName := fmt.Sprintf("%s_%s", strings.TrimSpace(sourceType), uploadID)
	zipPath := filepath.Join(baseTmpDir, folderName+".zip")
	extractDir := filepath.Join(baseTmpDir, folderName)

	if err := saveMultipartFile(file, zipPath); err != nil {
		return migrationModel.UploadMigrationSourceZipResponse{}, fmt.Errorf("save uploaded zip: %w", err)
	}
	defer func() {
		_ = os.Remove(zipPath)
	}()

	if err := os.MkdirAll(extractDir, 0o755); err != nil {
		return migrationModel.UploadMigrationSourceZipResponse{}, fmt.Errorf("create extract dir: %w", err)
	}
	if err := backup.UnpackZipToDir(zipPath, extractDir); err != nil {
		_ = os.RemoveAll(extractDir)
		return migrationModel.UploadMigrationSourceZipResponse{}, fmt.Errorf("unpack migration zip: %w", err)
	}

	relativeTmpDir := filepath.ToSlash(filepath.Join(migrationTmpRelativeDir, folderName))
	sourcePayload := map[string]any{
		"tmp_dir": relativeTmpDir,
	}
	return migrationModel.UploadMigrationSourceZipResponse{
		SourceType:    sourceType,
		TmpDir:        relativeTmpDir,
		SourcePayload: sourcePayload,
	}, nil
}

func (s *MigratorService) StartGlobalMigration(
	ctx context.Context,
	req migrationModel.StartGlobalMigrationRequest,
) (migrationModel.GlobalMigrationStateDTO, error) {
	if err := validateStartRequest(req); err != nil {
		return migrationModel.GlobalMigrationStateDTO{}, err
	}
	if _, err := s.ensureAdmin(ctx); err != nil {
		return migrationModel.GlobalMigrationStateDTO{}, err
	}

	s.activeMu.Lock()
	defer s.activeMu.Unlock()

	state, err := s.getGlobalState(ctx)
	if err != nil {
		return migrationModel.GlobalMigrationStateDTO{}, err
	}

	if state.Status != migrationModel.MigrationStatusIdle {
		return migrationModel.GlobalMigrationStateDTO{}, errors.New("请先结束/清理当前迁移")
	}

	now := nowUTC()
	state = migrationModel.GlobalMigrationStateDTO{
		Version:       1,
		SourceType:    strings.TrimSpace(req.SourceType),
		Status:        migrationModel.MigrationStatusPending,
		ErrorMessage:  "",
		SourcePayload: req.SourcePayload,
		StartedAt:     &now,
		UpdatedAt:     &now,
		FinishedAt:    nil,
	}
	if err := s.saveGlobalState(ctx, state); err != nil {
		return migrationModel.GlobalMigrationStateDTO{}, err
	}

	runCtx, cancel := context.WithCancel(context.Background())
	s.activeCancel = cancel
	go s.runGlobalMigration(runCtx, state)
	return state, nil
}

func (s *MigratorService) GetGlobalMigrationStatus(ctx context.Context) (migrationModel.GlobalMigrationStateDTO, error) {
	if _, err := s.ensureAdmin(ctx); err != nil {
		return migrationModel.GlobalMigrationStateDTO{}, err
	}
	return s.getGlobalState(ctx)
}

func (s *MigratorService) CancelGlobalMigration(ctx context.Context) (migrationModel.GlobalMigrationStateDTO, error) {
	if _, err := s.ensureAdmin(ctx); err != nil {
		return migrationModel.GlobalMigrationStateDTO{}, err
	}
	state, err := s.getGlobalState(ctx)
	if err != nil {
		return migrationModel.GlobalMigrationStateDTO{}, err
	}
	if state.Status != migrationModel.MigrationStatusPending && state.Status != migrationModel.MigrationStatusRunning {
		return migrationModel.GlobalMigrationStateDTO{}, errors.New(commonModel.INVALID_REQUEST_BODY)
	}
	now := nowUTC()
	state.Status = migrationModel.MigrationStatusCancelled
	state.ErrorMessage = "迁移已取消"
	state.UpdatedAt = &now
	state.FinishedAt = &now
	if err := s.saveGlobalState(ctx, state); err != nil {
		return migrationModel.GlobalMigrationStateDTO{}, err
	}
	s.activeMu.Lock()
	if s.activeCancel != nil {
		s.activeCancel()
		s.activeCancel = nil
	}
	s.activeMu.Unlock()
	return state, nil
}

func (s *MigratorService) CleanupGlobalMigration(ctx context.Context) error {
	if _, err := s.ensureAdmin(ctx); err != nil {
		return err
	}
	state, err := s.getGlobalState(ctx)
	if err != nil {
		return err
	}
	if state.Status == migrationModel.MigrationStatusPending || state.Status == migrationModel.MigrationStatusRunning {
		return errors.New("迁移进行中，无法清理")
	}
	return s.keyValueRepository.DeleteKeyValue(ctx, commonModel.MigrationGlobalJobStateKey)
}

func (s *MigratorService) runGlobalMigration(ctx context.Context, state migrationModel.GlobalMigrationStateDTO) {
	defer func() {
		s.activeMu.Lock()
		s.activeCancel = nil
		s.activeMu.Unlock()
	}()

	runner, err := coreMigrator.BuildSourceMigrator(state.SourceType)
	if err != nil {
		s.updateFailed(context.Background(), state, fmt.Sprintf("构建迁移器失败: %v", err))
		return
	}

	runningState := state
	now := nowUTC()
	runningState.Status = migrationModel.MigrationStatusRunning
	runningState.UpdatedAt = &now
	_ = s.saveGlobalState(context.Background(), runningState)

	result, runErr := runner.Migrate(ctx, spec.MigrateRequest{
		SourcePayload: runningState.SourcePayload,
		UpdateProgress: func(progress spec.MigrateProgress) {
			current, err := s.getGlobalState(context.Background())
			if err != nil {
				return
			}
			if current.Status == migrationModel.MigrationStatusCancelled {
				return
			}
			now := nowUTC()
			if strings.TrimSpace(progress.ErrorSummary) != "" {
				current.ErrorMessage = progress.ErrorSummary
			}
			current.UpdatedAt = &now
			_ = s.saveGlobalState(context.Background(), current)
		},
	})
	if runErr != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			return
		}
		s.updateFailed(context.Background(), runningState, runErr.Error())
		return
	}

	current, err := s.getGlobalState(context.Background())
	if err != nil {
		return
	}
	if current.Status == migrationModel.MigrationStatusCancelled {
		return
	}
	now = nowUTC()
	current.Status = migrationModel.MigrationStatusSuccess
	if strings.TrimSpace(result.ErrorSummary) != "" {
		current.ErrorMessage = result.ErrorSummary
	}
	current.UpdatedAt = &now
	current.FinishedAt = &now
	_ = s.saveGlobalState(context.Background(), current)
}

func (s *MigratorService) updateFailed(ctx context.Context, state migrationModel.GlobalMigrationStateDTO, reason string) {
	now := nowUTC()
	state.Status = migrationModel.MigrationStatusFailed
	state.ErrorMessage = reason
	state.UpdatedAt = &now
	state.FinishedAt = &now
	_ = s.saveGlobalState(ctx, state)
}

func validateStartRequest(req migrationModel.StartGlobalMigrationRequest) error {
	if err := validateSourceType(req.SourceType); err != nil {
		return err
	}
	tmpDir, ok := req.SourcePayload["tmp_dir"].(string)
	if !ok || strings.TrimSpace(tmpDir) == "" {
		return errors.New(commonModel.INVALID_REQUEST_BODY)
	}
	return nil
}

func (s *MigratorService) ensureAdmin(ctx context.Context) (string, error) {
	userID := viewer.MustFromContext(ctx).UserID()
	user, err := s.commonService.CommonGetUserByUserId(ctx, userID)
	if err != nil {
		return "", err
	}
	if !user.IsAdmin {
		return "", errors.New(commonModel.NO_PERMISSION_DENIED)
	}
	return userID, nil
}

func (s *MigratorService) getGlobalState(ctx context.Context) (migrationModel.GlobalMigrationStateDTO, error) {
	raw, err := s.keyValueRepository.GetKeyValue(ctx, commonModel.MigrationGlobalJobStateKey)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return migrationModel.GlobalMigrationStateDTO{
				Version:      1,
				Status:       migrationModel.MigrationStatusIdle,
				ErrorMessage: "",
			}, nil
		}
		return migrationModel.GlobalMigrationStateDTO{}, err
	}
	var state migrationModel.GlobalMigrationStateDTO
	if err := json.Unmarshal([]byte(raw), &state); err != nil {
		return migrationModel.GlobalMigrationStateDTO{}, err
	}
	if state.Version == 0 {
		state.Version = 1
	}
	if state.Status == "" {
		state.Status = migrationModel.MigrationStatusIdle
	}
	return state, nil
}

func (s *MigratorService) saveGlobalState(ctx context.Context, state migrationModel.GlobalMigrationStateDTO) error {
	state.Version = 1
	raw, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return s.keyValueRepository.AddOrUpdateKeyValue(ctx, commonModel.MigrationGlobalJobStateKey, string(raw))
}

func nowUTC() time.Time {
	return time.Now().UTC()
}

func validateSourceType(sourceType string) error {
	switch strings.TrimSpace(sourceType) {
	case migrationModel.MigrationSourceMemos, migrationModel.MigrationSourceEch0V3, migrationModel.MigrationSourceEch0V4:
		return nil
	default:
		return errors.New(commonModel.INVALID_REQUEST_BODY)
	}
}

func saveMultipartFile(file *multipart.FileHeader, dstPath string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer func() {
		_ = src.Close()
	}()

	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = dst.Close()
	}()

	_, err = io.Copy(dst, src)
	return err
}
