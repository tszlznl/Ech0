// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

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

	"github.com/lin-snow/ech0/internal/backup"
	"github.com/lin-snow/ech0/internal/job"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	jobModel "github.com/lin-snow/ech0/internal/model/job"
	migrationModel "github.com/lin-snow/ech0/internal/model/migration"
	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
	"github.com/lin-snow/ech0/pkg/viewer"
)

const migrationTmpRelativeDir = "files/tmp"

// MigratorService 是迁移领域服务的 HTTP 生命周期编排（start/status/cancel/cleanup）。
// 它委托给 job.Manager，保持对前端 API 契约不变（含 idle 哨兵）；实际导入由 Importer
// 承担（见 importer.go）。原手写状态机（KeyValue 单槽 + activeCancel）已删除。
type MigratorService struct {
	commonService CommonService
	jobManager    *job.Manager
}

func NewMigratorService(commonService CommonService, jobManager *job.Manager) *MigratorService {
	return &MigratorService{commonService: commonService, jobManager: jobManager}
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
	// 非 idle（存在任何作业行，无论在跑还是终态）则要求先清理，沿用旧语义。
	if _, err := s.jobManager.Get(ctx, jobModel.TypeMigration); err == nil {
		return migrationModel.UploadMigrationSourceZipResponse{}, errors.New("请先结束/清理当前迁移")
	} else if !errors.Is(err, job.ErrNotFound) {
		return migrationModel.UploadMigrationSourceZipResponse{}, err
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
	return migrationModel.UploadMigrationSourceZipResponse{
		SourceType:    sourceType,
		TmpDir:        relativeTmpDir,
		SourcePayload: map[string]any{"tmp_dir": relativeTmpDir},
	}, nil
}

// StartGlobalMigration 提交一次迁移作业（互斥 + 持久化 + goroutine 生命周期由 job.Manager 负责）。
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

	sourcePayload := cloneMap(req.SourcePayload)
	if _, ok := sourcePayload["created_by"]; !ok {
		sourcePayload["created_by"] = adminUserID
	}
	raw, err := json.Marshal(migrationModel.MigrationPayload{
		SourceType:    strings.TrimSpace(req.SourceType),
		SourcePayload: sourcePayload,
	})
	if err != nil {
		return migrationModel.GlobalMigrationStateDTO{}, err
	}

	jb, err := s.jobManager.Submit(ctx, jobModel.TypeMigration, raw)
	if err != nil {
		// 提交失败（如同类型互斥）：清理刚上传的 tmp，沿用旧「请先结束/清理当前迁移」。
		_ = cleanupMigrationTmpDirFromPayload(req.SourcePayload)
		if errors.Is(err, job.ErrAlreadyRunning) {
			return migrationModel.GlobalMigrationStateDTO{}, errors.New("请先结束/清理当前迁移")
		}
		return migrationModel.GlobalMigrationStateDTO{}, err
	}
	return s.jobToDTO(jb), nil
}

// GetGlobalMigrationStatus 查询当前状态；查无作业行时合成 idle 哨兵。
func (s *MigratorService) GetGlobalMigrationStatus(ctx context.Context) (migrationModel.GlobalMigrationStateDTO, error) {
	if _, err := s.ensureAdmin(ctx); err != nil {
		return migrationModel.GlobalMigrationStateDTO{}, err
	}
	jb, err := s.jobManager.Get(ctx, jobModel.TypeMigration)
	if errors.Is(err, job.ErrNotFound) {
		return migrationModel.GlobalMigrationStateDTO{Version: 1, Status: migrationModel.MigrationStatusIdle}, nil
	}
	if err != nil {
		return migrationModel.GlobalMigrationStateDTO{}, err
	}
	return s.jobToDTO(jb), nil
}

// CancelGlobalMigration 协作式取消在跑迁移；返回当前状态（前端轮询收敛到 cancelled）。
func (s *MigratorService) CancelGlobalMigration(ctx context.Context) (migrationModel.GlobalMigrationStateDTO, error) {
	if _, err := s.ensureAdmin(ctx); err != nil {
		return migrationModel.GlobalMigrationStateDTO{}, err
	}
	jb, err := s.jobManager.Get(ctx, jobModel.TypeMigration)
	if errors.Is(err, job.ErrNotFound) {
		return migrationModel.GlobalMigrationStateDTO{}, errors.New(commonModel.INVALID_REQUEST_BODY)
	}
	if err != nil {
		return migrationModel.GlobalMigrationStateDTO{}, err
	}
	if jb.Status != jobModel.StatusPending && jb.Status != jobModel.StatusRunning {
		return migrationModel.GlobalMigrationStateDTO{}, errors.New(commonModel.INVALID_REQUEST_BODY)
	}
	_ = s.jobManager.Cancel(jobModel.TypeMigration)
	jb, err = s.jobManager.Get(ctx, jobModel.TypeMigration)
	if err != nil {
		return migrationModel.GlobalMigrationStateDTO{}, err
	}
	return s.jobToDTO(jb), nil
}

// CleanupGlobalMigration 清理 tmp 目录并删除作业行（复位 idle）。
func (s *MigratorService) CleanupGlobalMigration(ctx context.Context) error {
	if _, err := s.ensureAdmin(ctx); err != nil {
		return err
	}
	jb, err := s.jobManager.Get(ctx, jobModel.TypeMigration)
	if errors.Is(err, job.ErrNotFound) {
		return nil // 已是 idle，幂等
	}
	if err != nil {
		return err
	}
	if jb.Status == jobModel.StatusPending || jb.Status == jobModel.StatusRunning {
		return errors.New("迁移进行中，无法清理")
	}
	var payload migrationModel.MigrationPayload
	if jb.Payload != "" {
		_ = json.Unmarshal([]byte(jb.Payload), &payload)
	}
	if err := cleanupMigrationTmpDirFromPayload(payload.SourcePayload); err != nil {
		return fmt.Errorf("cleanup migration tmp dir: %w", err)
	}
	return s.jobManager.Delete(ctx, jobModel.TypeMigration)
}

// jobToDTO 把通用 Job 映射回前端契约的 GlobalMigrationStateDTO（适配层，不污染框架）。
func (s *MigratorService) jobToDTO(jb jobModel.Job) migrationModel.GlobalMigrationStateDTO {
	var payload migrationModel.MigrationPayload
	if jb.Payload != "" {
		_ = json.Unmarshal([]byte(jb.Payload), &payload)
	}
	dto := migrationModel.GlobalMigrationStateDTO{
		Version:       1,
		SourceType:    payload.SourceType,
		Status:        string(jb.Status),
		Phase:         jb.Phase,
		ErrorMessage:  jb.Error,
		SourcePayload: payload.SourcePayload,
		StartedAt:     jb.StartedAt,
		FinishedAt:    jb.FinishedAt,
	}
	if jb.UpdatedAt != 0 {
		updatedAt := jb.UpdatedAt
		dto.UpdatedAt = &updatedAt
	}
	return dto
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

func validateSourceType(sourceType string) error {
	switch strings.TrimSpace(sourceType) {
	case migrationModel.MigrationSourceMemos, migrationModel.MigrationSourceEch0V4:
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
