package storage

import (
	"testing"

	"github.com/lin-snow/ech0/internal/config"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
)

func TestMergeStorageConfig_DBOverrideAndFallback(t *testing.T) {
	defaultCfg := config.StorageConfig{
		Mode:          "local",
		ObjectEnabled: false,
		DataRoot:      "data/files",
		Endpoint:      "fallback.example.com",
		AccessKey:     "fallback-ak",
		SecretKey:     "fallback-sk",
		BucketName:    "fallback-bucket",
		Region:        "ap-southeast-1",
		Provider:      "minio",
		UseSSL:        false,
		CDNURL:        "https://cdn.example.com/",
		PathPrefix:    "fallback/",
	}
	dbSetting := &settingModel.S3Setting{
		Enable:     true,
		Provider:   "aws",
		Endpoint:   "https://s3.amazonaws.com",
		BucketName: "db-bucket",
		UseSSL:     true,
		PathPrefix: "uploads/",
	}

	merged := MergeStorageConfig(defaultCfg, dbSetting)

	if !merged.ObjectEnabled {
		t.Fatalf("expected object storage enabled")
	}
	if merged.Mode != "object" {
		t.Fatalf("expected mode object, got %s", merged.Mode)
	}
	if merged.Provider != "aws" {
		t.Fatalf("expected provider aws, got %s", merged.Provider)
	}
	if merged.Endpoint != "s3.amazonaws.com" {
		t.Fatalf("expected trimmed endpoint, got %s", merged.Endpoint)
	}
	if merged.AccessKey != "fallback-ak" {
		t.Fatalf("expected access key fallback, got %s", merged.AccessKey)
	}
	if merged.BucketName != "db-bucket" {
		t.Fatalf("expected db bucket, got %s", merged.BucketName)
	}
	if merged.PathPrefix != "uploads" {
		t.Fatalf("expected trimmed path prefix, got %s", merged.PathPrefix)
	}
}

func TestStorageManager_ApplyS3Setting_InvalidConfigKeepsOldSelector(t *testing.T) {
	m := NewStorageManager(nil)
	before := m.GetSelector()
	if before == nil {
		t.Fatalf("expected initial selector")
	}

	err := m.ApplyS3Setting(settingModel.S3Setting{
		Enable:     true,
		Provider:   "minio",
		Endpoint:   "127.0.0.1:9000",
		AccessKey:  "",
		SecretKey:  "",
		BucketName: "",
		Region:     "",
		UseSSL:     false,
	})
	if err == nil {
		t.Fatalf("expected apply to fail for invalid object config")
	}

	after := m.GetSelector()
	if after != before {
		t.Fatalf("expected selector pointer unchanged after failed apply")
	}
	if after.ObjectEnabled() {
		t.Fatalf("expected object storage disabled after failed apply")
	}
}
