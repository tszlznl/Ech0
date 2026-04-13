package migration_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/lin-snow/ech0/internal/database"
	dbMigration "github.com/lin-snow/ech0/internal/database/migration"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestNormalizeLegacyStorageTimesToUTC_AppliesAndIdempotent(t *testing.T) {
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger:  logger.Default.LogMode(logger.Silent),
		NowFunc: func() time.Time { return time.Now().UTC() },
	})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	database.SetDB(db)
	if err := database.MigrateDB(); err != nil {
		t.Fatalf("migrate db failed: %v", err)
	}

	legacyTime := "2026-01-01 12:00:00"
	if err := db.Exec(
		"INSERT INTO comments (id, echo_id, nickname, email, content, status, hot, source, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"c1", "e1", "n1", "u@example.com", "hello", "approved", false, "guest", legacyTime, legacyTime,
	).Error; err != nil {
		t.Fatalf("insert legacy comment failed: %v", err)
	}

	report, err := dbMigration.NormalizeLegacyStorageTimesToUTC(db, "Asia/Shanghai")
	if err != nil {
		t.Fatalf("normalize legacy times failed: %v", err)
	}
	if report.TotalUpdated == 0 {
		t.Fatal("expected at least one updated row")
	}

	var createdAt string
	if err := db.Raw("SELECT created_at FROM comments WHERE id = ?", "c1").Scan(&createdAt).Error; err != nil {
		t.Fatalf("query normalized created_at failed: %v", err)
	}
	if !strings.HasPrefix(createdAt, "2026-01-01T04:00:00") {
		t.Fatalf("expected UTC shifted value 2026-01-01T04:00:00..., got %s", createdAt)
	}
}

func TestMigrate_IdempotentByMigratorKey(t *testing.T) {
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger:  logger.Default.LogMode(logger.Silent),
		NowFunc: func() time.Time { return time.Now().UTC() },
	})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	database.SetDB(db)
	if err := database.MigrateDB(); err != nil {
		t.Fatalf("migrate db failed: %v", err)
	}

	legacyTime := "2026-01-01 12:00:00"
	if err := db.Exec(
		"INSERT INTO comments (id, echo_id, nickname, email, content, status, hot, source, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"c2", "e1", "n1", "u@example.com", "hello", "approved", false, "guest", legacyTime, legacyTime,
	).Error; err != nil {
		t.Fatalf("insert legacy comment failed: %v", err)
	}

	dbMigration.Migrate(db, dbMigration.WithMigrators(dbMigration.NewLegacyTimeNormalizerMigrator(dbMigration.DefaultLegacySourceTimezone)))

	var first string
	if err := db.Raw("SELECT created_at FROM comments WHERE id = ?", "c2").Scan(&first).Error; err != nil {
		t.Fatalf("query first normalized created_at failed: %v", err)
	}
	if !strings.HasPrefix(first, "2026-01-01T04:00:00") {
		t.Fatalf("expected first normalized value 2026-01-01T04:00:00..., got %s", first)
	}

	dbMigration.Migrate(db, dbMigration.WithMigrators(dbMigration.NewLegacyTimeNormalizerMigrator(dbMigration.DefaultLegacySourceTimezone)))

	var second string
	if err := db.Raw("SELECT created_at FROM comments WHERE id = ?", "c2").Scan(&second).Error; err != nil {
		t.Fatalf("query second normalized created_at failed: %v", err)
	}
	if second != first {
		t.Fatalf("expected idempotent migration, first=%s second=%s", first, second)
	}

	var marker commonModel.KeyValue
	if err := db.Where("key = ?", commonModel.StorageTimeUTCNormalizedKey).First(&marker).Error; err != nil {
		t.Fatalf("expected migrator marker, got err: %v", err)
	}
}
