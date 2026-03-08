package storage

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"

	"github.com/lin-snow/ech0/internal/config"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
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
	return m
}

func (m *Manager) GetSelector() *StorageSelector {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.selector
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
	shouldEnableObject := cfg.ObjectEnabled || NormalizeStorageMode(cfg.Mode) == StorageModeObject
	if shouldEnableObject && !selector.ObjectEnabled() {
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
	if dbS3Setting.Enable {
		cfg.Mode = string(StorageModeObject)
	} else {
		cfg.Mode = string(StorageModeLocal)
	}

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
