package service

import "testing"

func TestParseMigratedS3Setting(t *testing.T) {
	report := map[string]any{
		"source_s3_setting": map[string]any{
			"enable":      true,
			"provider":    "r2",
			"endpoint":    "example.r2.cloudflarestorage.com",
			"access_key":  "ak",
			"secret_key":  "sk",
			"bucket_name": "bucket-a",
			"region":      "auto",
			"use_ssl":     true,
		},
	}
	setting, ok, err := parseMigratedS3Setting(report)
	if err != nil {
		t.Fatalf("parseMigratedS3Setting returned error: %v", err)
	}
	if !ok || setting == nil {
		t.Fatalf("expected valid migrated s3 setting")
	}
	if setting.Provider != "r2" || setting.BucketName != "bucket-a" {
		t.Fatalf("unexpected setting parsed: %+v", setting)
	}
}

func TestParseMigratedS3Setting_Invalid(t *testing.T) {
	report := map[string]any{
		"source_s3_setting": map[string]any{
			"provider": "r2",
		},
	}
	setting, ok, err := parseMigratedS3Setting(report)
	if err != nil {
		t.Fatalf("parseMigratedS3Setting returned error: %v", err)
	}
	if ok || setting != nil {
		t.Fatalf("expected invalid migrated s3 setting to be ignored")
	}
}
