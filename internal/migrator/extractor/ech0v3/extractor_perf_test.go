package ech0v3

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/lin-snow/ech0/internal/database"
	"github.com/lin-snow/ech0/internal/migrator/spec"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestEch0V3MigrateRegression(t *testing.T) {
	setRepoRootAsWorkingDir(t)

	targetDB := setupMigrationTargetDB(t)
	database.SetDB(targetDB)

	extractor := NewExtractor()
	start := time.Now()
	result, err := extractor.Migrate(context.Background(), spec.MigrateRequest{
		SourcePayload: map[string]any{
			"tmp_dir":     "files/tmp/ech0_v3_019cdc8c-b7ef-7278-9eb9-77185a34184e",
			"created_by":  "migration-test-user",
			"failure_threshold": 0.9,
		},
	})
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("migrate failed: %v", err)
	}

	t.Logf("migration elapsed=%s processed=%d success=%d fail=%d", elapsed, result.Processed, result.SuccessCount, result.FailCount)
	if result.Total == 0 {
		t.Fatalf("unexpected total=0")
	}
	if result.Processed != result.SuccessCount+result.FailCount {
		t.Fatalf("invalid processed stats: processed=%d success=%d fail=%d", result.Processed, result.SuccessCount, result.FailCount)
	}
	if result.SuccessCount == 0 {
		t.Fatalf("unexpected success_count=0")
	}
	if result.FailCount == result.Total {
		t.Fatalf("all rows failed unexpectedly")
	}

	var tagCount int64
	if err := targetDB.Model(&echoModel.Tag{}).Count(&tagCount).Error; err != nil {
		t.Fatalf("count tags failed: %v", err)
	}
	if tagCount == 0 {
		t.Fatalf("expected tags migrated, got 0")
	}

	var echoFileCount int64
	if err := targetDB.Model(&fileModel.EchoFile{}).Count(&echoFileCount).Error; err != nil {
		t.Fatalf("count echo_files failed: %v", err)
	}
	if echoFileCount == 0 {
		t.Fatalf("expected echo_files migrated, got 0")
	}
}

func setRepoRootAsWorkingDir(t *testing.T) {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("resolve test file path failed")
	}
	repoRoot := filepath.Clean(strings.TrimSuffix(filename, "internal/migrator/extractor/ech0v3/extractor_perf_test.go"))
	if err := os.Chdir(repoRoot); err != nil {
		t.Fatalf("chdir to repo root failed: %v", err)
	}
}

func setupMigrationTargetDB(t *testing.T) *gorm.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "migration_target.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open target db failed: %v", err)
	}
	if err := db.AutoMigrate(
		&echoModel.Echo{},
		&echoModel.EchoExtension{},
		&echoModel.Tag{},
		&echoModel.EchoTag{},
		&fileModel.File{},
		&fileModel.EchoFile{},
	); err != nil {
		t.Fatalf("auto migrate target db failed: %v", err)
	}
	return db
}

func TestBuildMigratedFileURL_NormalizeLegacyImagesURL(t *testing.T) {
	url := buildMigratedFileURL("1_1772108750_aa7e4e.jpeg", "/images/1_1772108750_aa7e4e.jpeg", "local")
	if url != "/api/files/images/1_1772108750_aa7e4e.jpeg" {
		t.Fatalf("unexpected migrated url: %s", url)
	}
}

func TestBuildMigratedFileURL_KeepSchemaPathWithoutDoublePrefix(t *testing.T) {
	url := buildMigratedFileURL("images/1_1772108750_aa7e4e.jpeg", "/images/1_1772108750_aa7e4e.jpeg", "local")
	if url != "/api/files/images/1_1772108750_aa7e4e.jpeg" {
		t.Fatalf("unexpected migrated url for schema path key: %s", url)
	}
}

func TestBuildObjectURLFromSetting(t *testing.T) {
	setting := &settingModel.S3Setting{
		Enable:     true,
		Endpoint:   "s3.example.com",
		BucketName: "legacy-bucket",
		PathPrefix: "upload",
		UseSSL:     true,
	}
	url := buildObjectURLFromSetting(*setting, "a.png")
	if url != "https://s3.example.com/legacy-bucket/upload/images/a.png" {
		t.Fatalf("unexpected s3 migrated url: %s", url)
	}
}

func TestMapSourceS3SettingToV4_Invalid(t *testing.T) {
	_, ok := mapSourceS3SettingToV4(&settingModel.S3Setting{
		Enable:     true,
		Provider:   "r2",
		BucketName: "bucket-only",
	})
	if ok {
		t.Fatalf("expected invalid source s3 setting")
	}
}

func TestNormalizeImageSource(t *testing.T) {
	if got := normalizeImageSource("s3", ""); got != "s3" {
		t.Fatalf("expected s3, got %s", got)
	}
	if got := normalizeImageSource("url", ""); got != "url" {
		t.Fatalf("expected url, got %s", got)
	}
	if got := normalizeImageSource("", "https://cdn.example.com/a.jpg"); got != "url" {
		t.Fatalf("expected url by absolute image_url, got %s", got)
	}
	if got := normalizeImageSource("", "/images/a.jpg"); got != "local" {
		t.Fatalf("expected local fallback, got %s", got)
	}
}
