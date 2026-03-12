package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/lin-snow/ech0/internal/backup"
	coreMigrator "github.com/lin-snow/ech0/internal/migrator"
	"github.com/lin-snow/ech0/internal/migrator/spec"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	migrationModel "github.com/lin-snow/ech0/internal/model/migration"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	echoRepository "github.com/lin-snow/ech0/internal/repository/echo"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
	"github.com/lin-snow/ech0/pkg/viewer"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const migrationTmpRelativeDir = "files/tmp"

type MigratorService struct {
	commonService      CommonService
	keyValueRepository KeyValueRepository
	storageManager     StorageManager
	appCache           AppCache

	activeMu     sync.Mutex
	activeCancel context.CancelFunc
}

func NewMigratorService(
	commonService CommonService,
	keyValueRepository KeyValueRepository,
	storageManager StorageManager,
	appCache AppCache,
) *MigratorService {
	return &MigratorService{
		commonService:      commonService,
		keyValueRepository: keyValueRepository,
		storageManager:     storageManager,
		appCache:           appCache,
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
	adminUserID, err := s.ensureAdmin(ctx)
	if err != nil {
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

	sourcePayload := cloneMap(req.SourcePayload)
	if _, ok := sourcePayload["created_by"]; !ok {
		sourcePayload["created_by"] = adminUserID
	}

	now := nowUTC()
	state = migrationModel.GlobalMigrationStateDTO{
		Version:       1,
		SourceType:    strings.TrimSpace(req.SourceType),
		Status:        migrationModel.MigrationStatusPending,
		ErrorMessage:  "",
		SourcePayload: sourcePayload,
		StartedAt:     &now,
		UpdatedAt:     &now,
		FinishedAt:    nil,
	}
	if err := s.saveGlobalStateWithRetry(ctx, state); err != nil {
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
	s.activeMu.Lock()
	cancelFn := s.activeCancel
	s.activeCancel = nil
	s.activeMu.Unlock()
	if cancelFn != nil {
		cancelFn()
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
	if err := s.saveGlobalStateWithRetry(ctx, state); err != nil {
		return migrationModel.GlobalMigrationStateDTO{}, err
	}
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
	if err := cleanupMigrationTmpDirFromPayload(state.SourcePayload); err != nil {
		return fmt.Errorf("cleanup migration tmp dir: %w", err)
	}
	return s.keyValueRepository.DeleteKeyValue(ctx, commonModel.MigrationGlobalJobStateKey)
}

func (s *MigratorService) runGlobalMigration(ctx context.Context, state migrationModel.GlobalMigrationStateDTO) {
	defer func() {
		s.activeMu.Lock()
		s.activeCancel = nil
		s.activeMu.Unlock()
	}()
	defer func() {
		if err := cleanupMigrationTmpDirFromPayload(state.SourcePayload); err != nil {
			logUtil.GetLogger().Warn("Failed to cleanup migration temp directory",
				zap.String("module", "migration"),
				zap.Error(err),
			)
		}
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
	_ = s.saveGlobalStateWithRetry(context.Background(), runningState)

	result, runErr := runner.Migrate(ctx, spec.MigrateRequest{
		SourcePayload: runningState.SourcePayload,
		UpdateProgress: func(progress spec.MigrateProgress) {
			if ctx.Err() != nil {
				return
			}
			if strings.TrimSpace(progress.ErrorSummary) != "" {
				runningState.ErrorMessage = progress.ErrorSummary
			}
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
	if result.Report != nil {
		current.SourcePayload["report"] = result.Report
	}
	if strings.TrimSpace(result.JobID) != "" {
		current.SourcePayload["migration_job_id"] = result.JobID
	}
	if strings.TrimSpace(result.ErrorSummary) != "" && result.FailCount > 0 {
		current.ErrorMessage = result.ErrorSummary
	}
	if err := s.applyMigratedSettings(context.Background(), result.Report); err != nil {
		s.updateFailed(context.Background(), runningState, fmt.Sprintf("应用迁移配置失败: %v", err))
		return
	}
	s.invalidateEchoCachesAfterMigration()
	current.UpdatedAt = &now
	current.FinishedAt = &now
	_ = s.saveGlobalStateWithRetry(context.Background(), current)
}

func (s *MigratorService) updateFailed(ctx context.Context, state migrationModel.GlobalMigrationStateDTO, reason string) {
	now := nowUTC()
	state.Status = migrationModel.MigrationStatusFailed
	state.ErrorMessage = reason
	state.UpdatedAt = &now
	state.FinishedAt = &now
	_ = s.saveGlobalStateWithRetry(ctx, state)
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

func (s *MigratorService) saveGlobalStateWithRetry(ctx context.Context, state migrationModel.GlobalMigrationStateDTO) error {
	var lastErr error
	for i := 0; i < 20; i++ {
		if err := s.saveGlobalState(ctx, state); err != nil {
			lastErr = err
			if !isDatabaseLockedError(err) {
				return err
			}
			select {
			case <-ctx.Done():
				return err
			default:
			}
			time.Sleep(50 * time.Millisecond)
			continue
		}
		return nil
	}
	return lastErr
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

func cloneMap(input map[string]any) map[string]any {
	if len(input) == 0 {
		return map[string]any{}
	}
	out := make(map[string]any, len(input))
	for k, v := range input {
		out[k] = v
	}
	return out
}

func isDatabaseLockedError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "database is locked")
}

func cleanupMigrationTmpDirFromPayload(sourcePayload map[string]any) error {
	tmpDir, ok := resolveMigrationTmpDir(sourcePayload)
	if !ok {
		return nil
	}
	return os.RemoveAll(tmpDir)
}

func resolveMigrationTmpDir(sourcePayload map[string]any) (string, bool) {
	if len(sourcePayload) == 0 {
		return "", false
	}
	tmpDirRaw, ok := sourcePayload["tmp_dir"].(string)
	if !ok || strings.TrimSpace(tmpDirRaw) == "" {
		return "", false
	}
	cleanRelPath := filepath.Clean(filepath.FromSlash(strings.TrimSpace(tmpDirRaw)))
	if cleanRelPath == "." || cleanRelPath == "" || filepath.IsAbs(cleanRelPath) || strings.HasPrefix(cleanRelPath, "..") {
		return "", false
	}

	allowedBaseDir := filepath.Clean(filepath.Join("data", migrationTmpRelativeDir))
	targetDir := filepath.Clean(filepath.Join("data", cleanRelPath))
	if targetDir != allowedBaseDir && !strings.HasPrefix(targetDir, allowedBaseDir+string(os.PathSeparator)) {
		return "", false
	}
	return targetDir, true
}

func (s *MigratorService) applyMigratedSettings(ctx context.Context, report map[string]any) error {
	if len(report) == 0 {
		return nil
	}
	updatedS3 := false

	if _, err := applyMigratedSettingValue(ctx, s.keyValueRepository, report, "source_system_setting", commonModel.SystemSettingsKey, parseMigratedSystemSetting); err != nil {
		return err
	}
	ok, err := applyMigratedSettingValue(ctx, s.keyValueRepository, report, "source_s3_setting", commonModel.S3SettingKey, parseMigratedS3Setting)
	if err != nil {
		return err
	}
	updatedS3 = ok
	if _, err := applyMigratedSettingValue(ctx, s.keyValueRepository, report, "source_oauth2_setting", commonModel.OAuth2SettingKey, parseMigratedOAuth2Setting); err != nil {
		return err
	}

	if updatedS3 && s.storageManager != nil {
		if err := s.storageManager.ReloadFromConfigAndDB(context.Background()); err != nil {
			return err
		}
	}
	return nil
}

func parseMigratedS3Setting(report map[string]any) (*settingModel.S3Setting, bool, error) {
	setting, ok, err := parseSettingFromReport[settingModel.S3Setting](report, "source_s3_setting")
	if err != nil || !ok || setting == nil {
		return nil, false, err
	}
	if strings.TrimSpace(setting.Provider) == "" || strings.TrimSpace(setting.Endpoint) == "" || strings.TrimSpace(setting.BucketName) == "" {
		return nil, false, nil
	}
	return setting, true, nil
}

func parseMigratedSystemSetting(report map[string]any) (*settingModel.SystemSetting, bool, error) {
	return parseSettingFromReport[settingModel.SystemSetting](report, "source_system_setting")
}

func parseMigratedOAuth2Setting(report map[string]any) (*settingModel.OAuth2Setting, bool, error) {
	return parseSettingFromReport[settingModel.OAuth2Setting](report, "source_oauth2_setting")
}

func applyMigratedSettingValue[T any](
	ctx context.Context,
	repo KeyValueRepository,
	report map[string]any,
	reportKey string,
	storeKey string,
	parser func(map[string]any) (*T, bool, error),
) (bool, error) {
	parsed, ok, err := parser(report)
	if err != nil {
		// 迁移报告中的单项配置格式异常时忽略，不中断整任务。
		return false, nil
	}
	if !ok || parsed == nil {
		return false, nil
	}
	raw, err := json.Marshal(parsed)
	if err != nil {
		return false, nil
	}
	if err := repo.AddOrUpdateKeyValue(ctx, storeKey, string(raw)); err != nil {
		return false, err
	}
	return reportKey == "source_s3_setting", nil
}

func parseSettingFromReport[T any](report map[string]any, key string) (*T, bool, error) {
	if len(report) == 0 {
		return nil, false, nil
	}
	raw, ok := report[key]
	if !ok || raw == nil {
		return nil, false, nil
	}
	bs, err := json.Marshal(raw)
	if err != nil {
		return nil, false, err
	}
	var setting T
	if err := json.Unmarshal(bs, &setting); err != nil {
		return nil, false, err
	}
	return &setting, true, nil
}

func (s *MigratorService) invalidateEchoCachesAfterMigration() {
	if s.appCache == nil {
		return
	}
	echoRepository.ClearEchoPageCache(s.appCache)
	echoRepository.ClearTodayEchosCache(s.appCache)
}
