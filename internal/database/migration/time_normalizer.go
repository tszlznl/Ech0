package migration

import (
	"fmt"
	"strings"
	"time"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const DefaultLegacySourceTimezone = "Asia/Shanghai"

type TimeColumnMigrationStat struct {
	Table        string
	Column       string
	CandidateRow int64
	UpdatedRow   int64
}

type TimeNormalizationReport struct {
	TimezoneSource string
	StartedAt      time.Time
	FinishedAt     time.Time
	TotalCandidate int64
	TotalUpdated   int64
	Details        []TimeColumnMigrationStat
}

type legacyTimeNormalizerMigrator struct {
	sourceTimezone string
}

func NewLegacyTimeNormalizerMigrator(sourceTimezone string) Migrator {
	return &legacyTimeNormalizerMigrator{sourceTimezone: strings.TrimSpace(sourceTimezone)}
}

func (m *legacyTimeNormalizerMigrator) Name() string {
	return "legacy_time_normalizer"
}

func (m *legacyTimeNormalizerMigrator) Key() string {
	return commonModel.StorageTimeUTCNormalizedKey
}

func (m *legacyTimeNormalizerMigrator) CanRerun() bool {
	return false
}

func (m *legacyTimeNormalizerMigrator) Migrate(db *gorm.DB) error {
	report, err := NormalizeLegacyStorageTimesToUTC(db, m.sourceTimezone)
	if err != nil {
		return err
	}

	logUtil.Info(
		"storage time normalization completed",
		zap.String("module", "database"),
		zap.String("migrator", m.Name()),
		zap.String("source_timezone", report.TimezoneSource),
		zap.Int64("total_candidate_rows", report.TotalCandidate),
		zap.Int64("total_updated_rows", report.TotalUpdated),
	)
	return nil
}

func NormalizeLegacyStorageTimesToUTC(db *gorm.DB, sourceTimezone string) (*TimeNormalizationReport, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	tzName := strings.TrimSpace(sourceTimezone)
	if tzName == "" {
		return nil, fmt.Errorf("source timezone is required")
	}
	loc, err := time.LoadLocation(tzName)
	if err != nil {
		return nil, fmt.Errorf("load timezone %s failed: %w", tzName, err)
	}
	_, offsetSeconds := time.Date(2000, 1, 1, 0, 0, 0, 0, loc).Zone()
	if offsetSeconds == 0 {
		return nil, fmt.Errorf("timezone %s offset is zero, migration would be a no-op", tzName)
	}

	report := &TimeNormalizationReport{
		TimezoneSource: tzName,
		StartedAt:      time.Now().UTC(),
		Details:        make([]TimeColumnMigrationStat, 0),
	}

	plans := []struct {
		table  string
		column string
	}{
		{"user_local_auth", "updated_at"},
		{"user_external_identities", "created_at"},
		{"user_external_identities", "updated_at"},
		{"webauthn_credentials", "last_used_at"},
		{"webauthn_credentials", "created_at"},
		{"webauthn_credentials", "updated_at"},
		{"echos", "created_at"},
		{"echo_extensions", "created_at"},
		{"echo_extensions", "updated_at"},
		{"tags", "created_at"},
		{"files", "created_at"},
		{"temp_files", "expire_at"},
		{"temp_files", "created_at"},
		{"comments", "created_at"},
		{"comments", "updated_at"},
		{"webhooks", "last_trigger"},
		{"webhooks", "created_at"},
		{"webhooks", "updated_at"},
		{"dead_letters", "next_retry"},
		{"dead_letters", "created_at"},
		{"dead_letters", "updated_at"},
		{"migration_jobs", "started_at"},
		{"migration_jobs", "finished_at"},
		{"migration_jobs", "created_at"},
		{"migration_jobs", "updated_at"},
		{"access_token_settings", "expiry"},
		{"access_token_settings", "last_used_at"},
		{"access_token_settings", "created_at"},
		{"passkeys", "last_used_at"},
		{"passkeys", "created_at"},
		{"passkeys", "updated_at"},
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		for _, plan := range plans {
			stat := TimeColumnMigrationStat{
				Table:  plan.table,
				Column: plan.column,
			}

			countSQL := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s IS NOT NULL", plan.table, plan.column)
			if err := tx.Raw(countSQL).Scan(&stat.CandidateRow).Error; err != nil {
				return fmt.Errorf("count %s.%s failed: %w", plan.table, plan.column, err)
			}

			updateSQL := fmt.Sprintf(
				"UPDATE %s SET %s = strftime('%%Y-%%m-%%dT%%H:%%M:%%fZ', julianday(%s) - (?/86400.0)) WHERE %s IS NOT NULL",
				plan.table,
				plan.column,
				plan.column,
				plan.column,
			)
			result := tx.Exec(updateSQL, offsetSeconds)
			if result.Error != nil {
				return fmt.Errorf("update %s.%s failed: %w", plan.table, plan.column, result.Error)
			}
			stat.UpdatedRow = result.RowsAffected
			report.TotalCandidate += stat.CandidateRow
			report.TotalUpdated += stat.UpdatedRow
			report.Details = append(report.Details, stat)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	report.FinishedAt = time.Now().UTC()
	return report, nil
}
