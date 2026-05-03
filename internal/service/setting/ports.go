// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"time"

	model "github.com/lin-snow/ech0/internal/model/setting"
	webhookModel "github.com/lin-snow/ech0/internal/model/webhook"
	commonService "github.com/lin-snow/ech0/internal/service/common"
	fileService "github.com/lin-snow/ech0/internal/service/file"
)

type Service interface {
	GetSetting(setting *model.SystemSetting) error
	UpdateSetting(ctx context.Context, newSetting *model.SystemSettingDto) error
	BootstrapDefaultLocale(ctx context.Context, locale string) error
	GetS3Setting(ctx context.Context, setting *model.S3Setting) error
	UpdateS3Setting(ctx context.Context, newSetting *model.S3SettingDto) error
	GetOAuth2Setting(ctx context.Context, setting *model.OAuth2Setting, forInternal bool) error
	UpdateOAuth2Setting(ctx context.Context, newSetting *model.OAuth2SettingDto) error
	GetOAuth2Status(status *model.OAuth2Status) error
	GetPasskeySetting(ctx context.Context, setting *model.PasskeySetting, forInternal bool) error
	UpdatePasskeySetting(ctx context.Context, newSetting *model.PasskeySettingDto) error
	GetPasskeyStatus(status *model.PasskeyStatus) error
	GetAllWebhooks(ctx context.Context) ([]webhookModel.Webhook, error)
	DeleteWebhook(ctx context.Context, id string) error
	UpdateWebhook(ctx context.Context, id string, newWebhook *model.WebhookDto) error
	CreateWebhook(ctx context.Context, newWebhook *model.WebhookDto) error
	TestWebhook(ctx context.Context, id string) error
	ListAccessTokens(ctx context.Context) ([]model.AccessTokenSetting, error)
	CreateAccessToken(ctx context.Context, newToken *model.AccessTokenSettingDto) (string, error)
	DeleteAccessToken(ctx context.Context, id string) error
	GetBackupScheduleSetting(setting *model.BackupSchedule) error
	UpdateBackupScheduleSetting(ctx context.Context, newSetting *model.BackupScheduleDto) error
	GetAgentInfo(setting *model.AgentSetting) error
	GetAgentSettings(ctx context.Context, setting *model.AgentSetting) error
	UpdateAgentSettings(ctx context.Context, newSetting *model.AgentSettingDto) error
}

type (
	CommonService = commonService.Service
	FileService   = fileService.Service
)

type KeyValueRepository interface {
	GetKeyValue(ctx context.Context, key string) (string, error)
	AddKeyValue(ctx context.Context, key, value string) error
	UpdateKeyValue(ctx context.Context, key, value string) error
	AddOrUpdateKeyValue(ctx context.Context, key, value string) error
	DeleteKeyValue(ctx context.Context, key string) error
}

type SettingRepository interface {
	ListAccessTokens(ctx context.Context, userID string) ([]model.AccessTokenSetting, error)
	CreateAccessToken(ctx context.Context, token *model.AccessTokenSetting) error
	GetAccessTokenByID(ctx context.Context, id string) (model.AccessTokenSetting, error)
	DeleteAccessTokenByID(ctx context.Context, id string) error
}

// TokenRevoker 写入 JTI 黑名单。SettingService 在管理员删除访问令牌时调用，
// 让 token 立刻失效而不是等到自然过期 (GHSA-fpw6-hrg5-q5x5)。
type TokenRevoker interface {
	RevokeToken(jti string, remainTTL time.Duration)
}

type WebhookRepository interface {
	GetAllWebhooks(ctx context.Context) ([]webhookModel.Webhook, error)
	GetWebhookByID(ctx context.Context, id string) (*webhookModel.Webhook, error)
	CreateWebhook(ctx context.Context, webhook *webhookModel.Webhook) error
	UpdateWebhookByID(ctx context.Context, id string, webhook *webhookModel.Webhook) error
	UpdateWebhookDeliveryStatus(ctx context.Context, id string, status string, lastTrigger int64) error
	DeleteWebhookByID(ctx context.Context, id string) error
}
