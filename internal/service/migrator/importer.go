// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	coreMigrator "github.com/lin-snow/ech0/internal/migrator"
	"github.com/lin-snow/ech0/internal/migrator/spec"
	commentModel "github.com/lin-snow/ech0/internal/model/comment"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	migrationModel "github.com/lin-snow/ech0/internal/model/migration"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	echoRepository "github.com/lin-snow/ech0/internal/repository/echo"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"go.uber.org/zap"
)

// Importer 跑迁移数据导入，不感知作业状态机（无 *job.Manager 依赖）。它与 MigratorService
// 拆开，是为了打破「MigratorService → job.Manager → MigrationRunner → MigratorService」
// 的构造环：MigrationRunner 只依赖 Importer，于是 Manager 可在构造期一次装配完成。
type Importer struct {
	durableKV      KVStore
	storageManager StorageManager
	appCache       AppCache
}

func NewImporter(
	durableKV KVStore,
	storageManager StorageManager,
	appCache AppCache,
) *Importer {
	return &Importer{
		durableKV:      durableKV,
		storageManager: storageManager,
		appCache:       appCache,
	}
}

// Import 选源迁移器 → Migrate（进度桥接到 report 阶段）→ 应用配置 → 失效缓存 →
// 清理 tmp。不写作业状态：失败/取消由 job.Manager 据返回 error 与 ctx.Err 落终态。
// 终态 result 为补充了 report/job_id 的 MigrationPayload，落 Job.Payload。
func (im *Importer) Import(
	ctx context.Context,
	payload migrationModel.MigrationPayload,
	report func(phase string, snapshot any),
) (any, error) {
	logUtil.GetLogger().Info("global migration started",
		zap.String("module", "migration"),
		zap.String("source_type", payload.SourceType),
	)
	defer func() {
		if err := cleanupMigrationTmpDirFromPayload(payload.SourcePayload); err != nil {
			logUtil.GetLogger().Warn("Failed to cleanup migration temp directory",
				zap.String("module", "migration"),
				zap.Error(err),
			)
		}
	}()

	runner, err := coreMigrator.BuildSourceMigrator(payload.SourceType)
	if err != nil {
		return nil, fmt.Errorf("构建迁移器失败: %v", err)
	}

	report(migrationModel.MigrationPhaseExtracting, nil)

	result, runErr := runner.Migrate(ctx, spec.MigrateRequest{
		SourcePayload: payload.SourcePayload,
		UpdateProgress: func(progress spec.MigrateProgress) {
			if ctx.Err() != nil {
				return
			}
			if phase := strings.TrimSpace(progress.CurrentPhase); phase != "" {
				report(phase, nil)
			}
		},
	})
	if runErr != nil {
		return nil, runErr
	}

	if err := im.applyMigratedSettings(context.Background(), result.Report); err != nil {
		return nil, fmt.Errorf("应用迁移配置失败: %v", err)
	}
	im.invalidateEchoCachesAfterMigration()
	report(migrationModel.MigrationPhaseCompleted, nil)

	logUtil.GetLogger().Info("global migration completed",
		zap.String("module", "migration"),
		zap.String("source_type", payload.SourceType),
		zap.Int64("processed", result.Processed),
		zap.Int64("success_count", result.SuccessCount),
		zap.Int64("fail_count", result.FailCount),
		zap.String("job_id", result.JobID),
	)

	enriched := payload
	if enriched.SourcePayload == nil {
		enriched.SourcePayload = map[string]any{}
	}
	if result.Report != nil {
		enriched.SourcePayload["report"] = result.Report
	}
	if strings.TrimSpace(result.JobID) != "" {
		enriched.SourcePayload["migration_job_id"] = result.JobID
	}
	return enriched, nil
}

func (im *Importer) applyMigratedSettings(ctx context.Context, report map[string]any) error {
	if len(report) == 0 {
		return nil
	}

	if _, err := applyMigratedSettingValue(ctx, im.durableKV, report, "source_system_setting", commonModel.SystemSettingsKey, parseMigratedSystemSetting); err != nil {
		return err
	}
	if _, err := applyMigratedSettingValue(ctx, im.durableKV, report, "source_comment_setting", commentModel.CommentSystemSettingKey, parseMigratedCommentSetting); err != nil {
		return err
	}
	updatedS3, err := applyMigratedSettingValue(ctx, im.durableKV, report, "source_s3_setting", commonModel.S3SettingKey, parseMigratedS3Setting)
	if err != nil {
		return err
	}
	if _, err := applyMigratedSettingValue(ctx, im.durableKV, report, "source_oauth2_setting", commonModel.OAuth2SettingKey, parseMigratedOAuth2Setting); err != nil {
		return err
	}

	if updatedS3 && im.storageManager != nil {
		if err := im.storageManager.ReloadFromConfigAndDB(context.Background()); err != nil {
			return err
		}
	}
	return nil
}

func (im *Importer) invalidateEchoCachesAfterMigration() {
	if im.appCache == nil {
		return
	}
	echoRepository.ClearEchoPageCache(im.appCache)
	echoRepository.ClearTodayEchosCache(im.appCache)
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

func parseMigratedCommentSetting(report map[string]any) (*commentModel.SystemSetting, bool, error) {
	return parseSettingFromReport[commentModel.SystemSetting](report, "source_comment_setting")
}

func parseMigratedOAuth2Setting(report map[string]any) (*settingModel.OAuth2Setting, bool, error) {
	return parseSettingFromReport[settingModel.OAuth2Setting](report, "source_oauth2_setting")
}

func applyMigratedSettingValue[T any](
	ctx context.Context,
	repo KVStore,
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
	if err := repo.Set(ctx, storeKey, string(raw)); err != nil {
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
