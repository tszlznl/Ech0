package migration_test

import (
	"fmt"
	"testing"

	"github.com/lin-snow/ech0/internal/database"
	dbMigration "github.com/lin-snow/ech0/internal/database/migration"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestOAuthBindingsDropMigrator_DropTableAndMarkDone(t *testing.T) {
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

	if err := db.Exec(`CREATE TABLE oauth_bindings (id TEXT PRIMARY KEY, user_id TEXT)`).Error; err != nil {
		t.Fatalf("create oauth_bindings failed: %v", err)
	}
	if err := db.Exec(`INSERT INTO oauth_bindings (id, user_id) VALUES ('b1', 'u1')`).Error; err != nil {
		t.Fatalf("insert oauth_bindings row failed: %v", err)
	}

	dbMigration.Migrate(
		db,
		dbMigration.WithStopOnError(),
		dbMigration.WithMigrators(dbMigration.NewOAuthBindingsDropMigrator()),
	)

	var exists int64
	if err := db.Raw("SELECT COUNT(1) FROM sqlite_master WHERE type='table' AND name='oauth_bindings'").Scan(&exists).Error; err != nil {
		t.Fatalf("query sqlite_master failed: %v", err)
	}
	if exists != 0 {
		t.Fatal("expected oauth_bindings to be dropped")
	}

	var marker commonModel.KeyValue
	if err := db.Where("key = ?", commonModel.OAuthBindingsDroppedKey).First(&marker).Error; err != nil {
		t.Fatalf("expected migrator marker, got err: %v", err)
	}
}

func TestLegacyInboxesDropMigrator_DropTableAndMarkDone(t *testing.T) {
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

	if err := db.Exec(`CREATE TABLE inboxes (id TEXT PRIMARY KEY)`).Error; err != nil {
		t.Fatalf("create inboxes failed: %v", err)
	}
	if err := db.Exec(`INSERT INTO inboxes (id) VALUES ('i1')`).Error; err != nil {
		t.Fatalf("insert inboxes row failed: %v", err)
	}

	dbMigration.Migrate(
		db,
		dbMigration.WithStopOnError(),
		dbMigration.WithMigrators(dbMigration.NewLegacyInboxesDropMigrator()),
	)

	var exists int64
	if err := db.Raw("SELECT COUNT(1) FROM sqlite_master WHERE type='table' AND name='inboxes'").Scan(&exists).Error; err != nil {
		t.Fatalf("query sqlite_master failed: %v", err)
	}
	if exists != 0 {
		t.Fatal("expected inboxes to be dropped")
	}

	var marker commonModel.KeyValue
	if err := db.Where("key = ?", commonModel.LegacyInboxesDroppedKey).First(&marker).Error; err != nil {
		t.Fatalf("expected migrator marker, got err: %v", err)
	}
}
