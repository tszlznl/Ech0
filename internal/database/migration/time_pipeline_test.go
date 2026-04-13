package migration_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/lin-snow/ech0/internal/database"
	dbMigration "github.com/lin-snow/ech0/internal/database/migration"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupMemoryDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	database.SetDB(db)
	if err := database.MigrateDB(); err != nil {
		t.Fatalf("migrate db failed: %v", err)
	}
	return db
}

func TestTimeMigrationPipeline_SanitizeValidateConvertAndRebuild(t *testing.T) {
	db := setupMemoryDB(t)

	dirtyValue := " 2026/01/01T12:00:00Z "
	if err := db.Exec(
		"INSERT INTO comments (id, echo_id, nickname, email, content, status, hot, source, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"pipeline-ok", "e1", "n1", "u@example.com", "hello", "approved", false, "guest", dirtyValue, dirtyValue,
	).Error; err != nil {
		t.Fatalf("insert dirty comment failed: %v", err)
	}

	dbMigration.Migrate(
		db,
		dbMigration.WithStopOnError(),
		dbMigration.WithMigrators(
			dbMigration.NewStorageTimeSanitizeMigrator(),
			dbMigration.NewStorageTimeValidateMigrator(),
			dbMigration.NewStorageTimeUnixMigrator(),
			dbMigration.NewStorageTimeSchemaRebuildMigrator(),
		),
	)

	var createdAtType string
	if err := db.Raw("SELECT typeof(created_at) FROM comments WHERE id = ?", "pipeline-ok").Scan(&createdAtType).Error; err != nil {
		t.Fatalf("query created_at typeof failed: %v", err)
	}
	if createdAtType != "integer" {
		t.Fatalf("expected created_at value type integer, got %s", createdAtType)
	}

	var createdAt int64
	if err := db.Raw("SELECT created_at FROM comments WHERE id = ?", "pipeline-ok").Scan(&createdAt).Error; err != nil {
		t.Fatalf("query created_at failed: %v", err)
	}
	if createdAt == 0 {
		t.Fatal("expected created_at unix timestamp > 0")
	}

	var pragmaRows []struct {
		Name string `gorm:"column:name"`
		Type string `gorm:"column:type"`
	}
	if err := db.Raw(`PRAGMA table_info("comments")`).Scan(&pragmaRows).Error; err != nil {
		t.Fatalf("query table_info comments failed: %v", err)
	}
	var declaredType string
	for _, row := range pragmaRows {
		if row.Name == "created_at" {
			declaredType = row.Type
			break
		}
	}
	if !strings.Contains(strings.ToUpper(declaredType), "INT") {
		t.Fatalf("expected comments.created_at declared type to contain INT, got %s", declaredType)
	}
}

func TestTimeMigrationPipeline_StopOnValidateFailure(t *testing.T) {
	db := setupMemoryDB(t)

	invalidValue := "not-a-valid-time"
	if err := db.Exec(
		"INSERT INTO comments (id, echo_id, nickname, email, content, status, hot, source, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"pipeline-fail", "e1", "n1", "u@example.com", "hello", "approved", false, "guest", invalidValue, invalidValue,
	).Error; err != nil {
		t.Fatalf("insert invalid comment failed: %v", err)
	}

	dbMigration.Migrate(
		db,
		dbMigration.WithStopOnError(),
		dbMigration.WithMigrators(
			dbMigration.NewStorageTimeSanitizeMigrator(),
			dbMigration.NewStorageTimeValidateMigrator(),
			dbMigration.NewStorageTimeUnixMigrator(),
		),
	)

	var createdAtType string
	if err := db.Raw("SELECT typeof(created_at) FROM comments WHERE id = ?", "pipeline-fail").Scan(&createdAtType).Error; err != nil {
		t.Fatalf("query created_at typeof failed: %v", err)
	}
	if createdAtType != "text" {
		t.Fatalf("expected created_at to remain text when validation fails, got %s", createdAtType)
	}

	var marker commonModel.KeyValue
	err := db.Where("key = ?", commonModel.StorageTimeUnixMigratedKey).First(&marker).Error
	if err == nil {
		t.Fatal("expected unix migrator marker absent when validate fails")
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("query unix migrator marker failed: %v", err)
	}
}
