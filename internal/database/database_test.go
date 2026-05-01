// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package database

import (
	"testing"
	"time"

	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestBuildGormConfig_UsesUTCNowFunc(t *testing.T) {
	cfg := buildGormConfig(logger.Silent)
	if cfg.NowFunc != nil {
		t.Fatalf("expected nil NowFunc, got configured function")
	}
}

func TestMigrateDB_AccessTokenSettingIncludesScopeFields(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	SetDB(db)

	if err := MigrateDB(); err != nil {
		t.Fatalf("migrate db failed: %v", err)
	}

	expiry := time.Now().UTC().Add(8 * time.Hour).Unix()
	lastUsed := time.Now().UTC().Unix()
	record := settingModel.AccessTokenSetting{
		UserID:     "user-1",
		Token:      "token-1",
		Name:       "test-token",
		TokenType:  "access",
		Scopes:     `["echo:read"]`,
		Audience:   "public-client",
		JTI:        "jti-1",
		Expiry:     &expiry,
		LastUsedAt: &lastUsed,
		CreatedAt:  time.Now().UTC().Unix(),
	}

	if err := GetDB().Create(&record).Error; err != nil {
		t.Fatalf("create access token failed: %v", err)
	}

	var got settingModel.AccessTokenSetting
	if err := GetDB().Where("id = ?", record.ID).First(&got).Error; err != nil {
		t.Fatalf("query access token failed: %v", err)
	}
	if got.TokenType != "access" {
		t.Fatalf("expected token type access, got %s", got.TokenType)
	}
	if got.Scopes == "" {
		t.Fatal("expected scopes not empty")
	}
	if got.Audience != "public-client" {
		t.Fatalf("expected audience public-client, got %s", got.Audience)
	}
	if got.JTI == "" {
		t.Fatal("expected jti not empty")
	}
	if got.LastUsedAt == nil {
		t.Fatal("expected last_used_at to be set")
	}
}
