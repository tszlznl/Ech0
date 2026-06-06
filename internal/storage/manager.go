// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package storage

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"

	"github.com/lin-snow/ech0/internal/config"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
)

type S3SettingStore interface {
	GetKeyValue(ctx context.Context, key string) (string, error)
}

type Manager struct {
	mu         sync.RWMutex
	defaultCfg config.StorageConfig
	store      S3SettingStore
	selector   *StorageSelector
}

func NewStorageManager(store S3SettingStore) *Manager {
	defaultCfg := config.Config().Storage
	m := &Manager{
		defaultCfg: defaultCfg,
		store:      store,
		selector:   NewStorageSelector(defaultCfg),
	}
	_ = m.ReloadFromConfigAndDB(context.Background())
	// Publish the live URL resolver to the File model so reads recompute URLs
	// from the current config instead of trusting the write-time snapshot.
	fileModel.RegisterURLResolver(m.ResolveURL)
	return m
}

func (m *Manager) GetSelector() *StorageSelector {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.selector
}

// ResolveURL builds a file's public URL from its stored storage type + key
// using the current selector. The File model's AfterFind hook calls this to
// refresh URL snapshots on read; it always reflects the latest CDN/S3 config.
func (m *Manager) ResolveURL(storageType, key string) string {
	return m.GetSelector().ResolveURL(StorageType(storageType), key)
}

// GetStorageConfig returns the current merged storage configuration.
func (m *Manager) GetStorageConfig(ctx context.Context) config.StorageConfig {
	dbSetting, _ := m.loadS3SettingFromDB(ctx)
	return MergeStorageConfig(m.defaultCfg, dbSetting)
}

func (m *Manager) ReloadFromConfigAndDB(ctx context.Context) error {
	dbSetting, err := m.loadS3SettingFromDB(ctx)
	if err != nil {
		return err
	}
	cfg := MergeStorageConfig(m.defaultCfg, dbSetting)
	return m.replaceSelector(cfg)
}

func (m *Manager) ApplyS3Setting(setting settingModel.S3Setting) error {
	cfg := MergeStorageConfig(m.defaultCfg, &setting)
	return m.replaceSelector(cfg)
}

func (m *Manager) replaceSelector(cfg config.StorageConfig) error {
	selector := NewStorageSelector(cfg)
	if cfg.ObjectEnabled && !selector.ObjectEnabled() {
		return errors.New("object storage enabled but initialization failed")
	}
	m.mu.Lock()
	m.selector = selector
	m.mu.Unlock()
	return nil
}

func (m *Manager) loadS3SettingFromDB(ctx context.Context) (*settingModel.S3Setting, error) {
	if m.store == nil {
		return nil, nil
	}
	raw, err := m.store.GetKeyValue(ctx, commonModel.S3SettingKey)
	if err != nil {
		return nil, nil
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	var s settingModel.S3Setting
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// MergeStorageConfig 把 env 基线（defaultCfg，来自 config.StorageConfig）与面板存库的
// S3 设置逐字段合并成 Manager 实际运行用的配置。
//
// 为何要在库设置之上叠加 env，而不是直接用库里的值：
//   - env（ECH0_S3_*）是一等的部署期配置通道。Manager 是 DI 构造期就要拉起存储后端的
//     单例（NewStorageManager 里立即 ReloadFromConfigAndDB），而 setting.Seed 要到
//     BeforeStart 才把 s3_setting 落库——构造早于 seeder，全新部署时库里还没这条记录，
//     只能靠 env 把后端跑起来；headless/docker 部署更是常年只配 env、从不开面板。
//   - 因此库里的 s3_setting 被当成 env 之上的「可选部分覆盖层」：coalesceTrim 让字段
//     有值用库值、留空回退 env。这支持「密钥放 env、可调项放面板」，也为 S3Setting 日后
//     新增字段提供前向兼容（旧记录缺该字段时回退 env，而非置空）。
//
// 注意：正因逐字段回退，env 里设过的字段无法仅靠面板清空——会被 env 重新填回，要真正
// 清掉得同时改 env。
func MergeStorageConfig(defaultCfg config.StorageConfig, dbS3Setting *settingModel.S3Setting) config.StorageConfig {
	cfg := defaultCfg
	cfg.DataRoot = strings.TrimSpace(defaultCfg.DataRoot)
	if cfg.DataRoot == "" {
		cfg.DataRoot = "data/files"
	}

	if dbS3Setting == nil {
		return cfg
	}

	cfg.ObjectEnabled = dbS3Setting.Enable

	cfg.Provider = coalesceTrim(dbS3Setting.Provider, cfg.Provider)
	cfg.Endpoint = trimEndpoint(coalesceTrim(dbS3Setting.Endpoint, cfg.Endpoint))
	cfg.AccessKey = coalesceTrim(dbS3Setting.AccessKey, cfg.AccessKey)
	cfg.SecretKey = coalesceTrim(dbS3Setting.SecretKey, cfg.SecretKey)
	cfg.BucketName = coalesceTrim(dbS3Setting.BucketName, cfg.BucketName)
	cfg.Region = coalesceTrim(dbS3Setting.Region, cfg.Region)
	cfg.CDNURL = strings.TrimRight(coalesceTrim(dbS3Setting.CDNURL, cfg.CDNURL), "/")
	cfg.PathPrefix = strings.Trim(strings.TrimSpace(coalesceTrim(dbS3Setting.PathPrefix, cfg.PathPrefix)), "/")
	cfg.UseSSL = dbS3Setting.UseSSL

	return cfg
}

func coalesceTrim(preferred string, fallback string) string {
	v := strings.TrimSpace(preferred)
	if v != "" {
		return v
	}
	return strings.TrimSpace(fallback)
}

func trimEndpoint(endpoint string) string {
	e := strings.TrimSpace(endpoint)
	e = strings.TrimPrefix(e, "http://")
	e = strings.TrimPrefix(e, "https://")
	return e
}
