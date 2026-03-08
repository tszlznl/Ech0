package service

import (
	"context"

	model "github.com/lin-snow/ech0/internal/model/setting"
	webhookModel "github.com/lin-snow/ech0/internal/model/webhook"
	commonService "github.com/lin-snow/ech0/internal/service/common"
)

type Service interface {
	GetSetting(setting *model.SystemSetting) error
	UpdateSetting(userid string, newSetting *model.SystemSettingDto) error
	GetCommentSetting(setting *model.CommentSetting) error
	UpdateCommentSetting(userid string, newSetting *model.CommentSettingDto) error
	GetS3Setting(userid string, setting *model.S3Setting) error
	UpdateS3Setting(userid string, newSetting *model.S3SettingDto) error
	GetOAuth2Setting(userid string, setting *model.OAuth2Setting, forInternal bool) error
	UpdateOAuth2Setting(userid string, newSetting *model.OAuth2SettingDto) error
	GetOAuth2Status(status *model.OAuth2Status) error
	GetAllWebhooks(userid string) ([]webhookModel.Webhook, error)
	DeleteWebhook(userid, id string) error
	UpdateWebhook(userid, id string, newWebhook *model.WebhookDto) error
	CreateWebhook(userid string, newWebhook *model.WebhookDto) error
	ListAccessTokens(userid string) ([]model.AccessTokenSetting, error)
	CreateAccessToken(userid string, newToken *model.AccessTokenSettingDto) (string, error)
	DeleteAccessToken(userid, id string) error
	GetBackupScheduleSetting(setting *model.BackupSchedule) error
	UpdateBackupScheduleSetting(userid string, newSetting *model.BackupScheduleDto) error
	GetAgentInfo(setting *model.AgentSetting) error
	GetAgentSettings(userid string, setting *model.AgentSetting) error
	UpdateAgentSettings(userid string, newSetting *model.AgentSettingDto) error
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
	ListAccessTokens(ctx context.Context, userID string) ([]model.AccessTokenSetting, error)
	CreateAccessToken(ctx context.Context, token *model.AccessTokenSetting) error
	DeleteAccessTokenByID(ctx context.Context, id string) error
}

type WebhookRepository interface {
	GetAllWebhooks(ctx context.Context) ([]webhookModel.Webhook, error)
	CreateWebhook(ctx context.Context, webhook *webhookModel.Webhook) error
	DeleteWebhookByID(ctx context.Context, id string) error
}
