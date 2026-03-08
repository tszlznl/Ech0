package service

import (
	"context"

	model "github.com/lin-snow/ech0/internal/model/setting"
	webhookModel "github.com/lin-snow/ech0/internal/model/webhook"
	commonService "github.com/lin-snow/ech0/internal/service/common"
)

type Service interface {
	GetSetting(setting *model.SystemSetting) error
	UpdateSetting(userid uint, newSetting *model.SystemSettingDto) error
	GetCommentSetting(setting *model.CommentSetting) error
	UpdateCommentSetting(userid uint, newSetting *model.CommentSettingDto) error
	GetS3Setting(userid uint, setting *model.S3Setting) error
	UpdateS3Setting(userid uint, newSetting *model.S3SettingDto) error
	GetOAuth2Setting(userid uint, setting *model.OAuth2Setting, forInternal bool) error
	UpdateOAuth2Setting(userid uint, newSetting *model.OAuth2SettingDto) error
	GetOAuth2Status(status *model.OAuth2Status) error
	GetAllWebhooks(userid uint) ([]webhookModel.Webhook, error)
	DeleteWebhook(userid, id uint) error
	UpdateWebhook(userid, id uint, newWebhook *model.WebhookDto) error
	CreateWebhook(userid uint, newWebhook *model.WebhookDto) error
	ListAccessTokens(userid uint) ([]model.AccessTokenSetting, error)
	CreateAccessToken(userid uint, newToken *model.AccessTokenSettingDto) (string, error)
	DeleteAccessToken(userid, id uint) error
	GetBackupScheduleSetting(setting *model.BackupSchedule) error
	UpdateBackupScheduleSetting(userid uint, newSetting *model.BackupScheduleDto) error
	GetAgentInfo(setting *model.AgentSetting) error
	GetAgentSettings(userid uint, setting *model.AgentSetting) error
	UpdateAgentSettings(userid uint, newSetting *model.AgentSettingDto) error
}

type CommonService = commonService.Service

type KeyValueRepository interface {
	GetKeyValue(ctx context.Context, key string) (string, error)
	AddKeyValue(ctx context.Context, key, value string) error
	UpdateKeyValue(ctx context.Context, key, value string) error
	AddOrUpdateKeyValue(ctx context.Context, key, value string) error
	DeleteKeyValue(ctx context.Context, key string) error
}

type SettingRepository interface {
	ListAccessTokens(ctx context.Context, userID uint) ([]model.AccessTokenSetting, error)
	CreateAccessToken(ctx context.Context, token *model.AccessTokenSetting) error
	DeleteAccessTokenByID(ctx context.Context, id uint) error
}

type WebhookRepository interface {
	GetAllWebhooks(ctx context.Context) ([]webhookModel.Webhook, error)
	CreateWebhook(ctx context.Context, webhook *webhookModel.Webhook) error
	DeleteWebhookByID(ctx context.Context, id uint) error
}
