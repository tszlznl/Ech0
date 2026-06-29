// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package handler 暴露系统设置相关的 HTTP 接口（Huma type-first，全部 JSON）。
package handler

import (
	"context"

	"github.com/lin-snow/ech0/internal/handler/humares"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
	webhookModel "github.com/lin-snow/ech0/internal/model/webhook"
	service "github.com/lin-snow/ech0/internal/service/setting"
)

type SettingHandler struct {
	settingService service.Service
}

// NewSettingHandler SettingHandler 的构造函数
func NewSettingHandler(settingService service.Service) *SettingHandler {
	return &SettingHandler{
		settingService: settingService,
	}
}

type (
	EmptyInput struct{}
	IDInput    struct {
		ID string `path:"id" format:"uuid" doc:"资源 ID（UUID）"`
	}
	UpdateSettingsInput struct{ Body model.SystemSettingDto }
	S3SettingInput      struct{ Body model.S3SettingDto }
	OAuth2SettingInput  struct{ Body model.OAuth2SettingDto }
	PasskeySettingInput struct{ Body model.PasskeySettingDto }
	WebhookInput        struct{ Body model.WebhookDto }
	WebhookIDBodyInput  struct {
		ID   string `path:"id" format:"uuid" doc:"Webhook ID（UUID）"`
		Body model.WebhookDto
	}
	AccessTokenInput      struct{ Body model.AccessTokenSettingDto }
	SnapshotScheduleInput struct{ Body model.SnapshotScheduleDto }
	AgentSettingInput     struct{ Body model.AgentSettingDto }
	EmbeddingSettingInput struct{ Body model.EmbeddingSettingDto }
)

// --- 公开 ---

// GetSettings 获取系统全局设置（公开）。
func (h *SettingHandler) GetSettings(ctx context.Context, _ *EmptyInput) (*humares.Envelope[model.SystemSetting], error) {
	var settings model.SystemSetting
	if err := h.settingService.GetSetting(&settings); err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK(ctx, settings, commonModel.GET_SETTINGS_SUCCESS), nil
}

// GetOAuth2Status 获取 OAuth2 启用状态（公开）。
func (h *SettingHandler) GetOAuth2Status(ctx context.Context, _ *EmptyInput) (*humares.Envelope[model.OAuth2Status], error) {
	var status model.OAuth2Status
	if err := h.settingService.GetOAuth2Status(&status); err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK(ctx, status, commonModel.GET_OAUTH2_STATUS_SUCCESS), nil
}

// GetPasskeyStatus 获取 Passkey 就绪状态（公开）。
func (h *SettingHandler) GetPasskeyStatus(ctx context.Context, _ *EmptyInput) (*humares.Envelope[model.PasskeyStatus], error) {
	var status model.PasskeyStatus
	if err := h.settingService.GetPasskeyStatus(&status); err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK(ctx, status, commonModel.GET_PASSKEY_STATUS_SUCCESS), nil
}

// GetAgentInfo 获取 Agent 公开信息（公开，敏感字段已脱敏）。
func (h *SettingHandler) GetAgentInfo(ctx context.Context, _ *EmptyInput) (*humares.Envelope[model.AgentSetting], error) {
	var settings model.AgentSetting
	if err := h.settingService.GetAgentInfo(&settings); err != nil {
		return nil, humares.Err(ctx, err)
	}
	settings.ApiKey = ""
	settings.Prompt = ""
	settings.BaseURL = ""
	return humares.OK(ctx, settings, commonModel.GET_SETTINGS_SUCCESS), nil
}

// --- admin:settings ---

// UpdateSettings 更新系统全局设置。
func (h *SettingHandler) UpdateSettings(ctx context.Context, in *UpdateSettingsInput) (*humares.Envelope[any], error) {
	if err := h.settingService.UpdateSetting(ctx, &in.Body); err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK[any](ctx, nil, commonModel.UPDATE_SETTINGS_SUCCESS), nil
}

// GetS3Settings 获取 S3 存储设置。
func (h *SettingHandler) GetS3Settings(ctx context.Context, _ *EmptyInput) (*humares.Envelope[model.S3Setting], error) {
	var s3Setting model.S3Setting
	if err := h.settingService.GetS3Setting(ctx, &s3Setting); err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK(ctx, s3Setting, commonModel.GET_S3_SETTINGS_SUCCESS), nil
}

// UpdateS3Settings 更新 S3 存储设置。
func (h *SettingHandler) UpdateS3Settings(ctx context.Context, in *S3SettingInput) (*humares.Envelope[any], error) {
	if err := h.settingService.UpdateS3Setting(ctx, &in.Body); err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK[any](ctx, nil, commonModel.UPDATE_S3_SETTINGS_SUCCESS), nil
}

// TestS3Connection 用提交的 S3 配置做一次连通性探测（不保存）。
func (h *SettingHandler) TestS3Connection(ctx context.Context, in *S3SettingInput) (*humares.Envelope[any], error) {
	if err := h.settingService.TestS3Connection(ctx, &in.Body); err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK[any](ctx, nil, commonModel.TEST_S3_CONNECTION_SUCCESS), nil
}

// GetOAuth2Settings 获取 OAuth2 设置。
func (h *SettingHandler) GetOAuth2Settings(ctx context.Context, _ *EmptyInput) (*humares.Envelope[model.OAuth2Setting], error) {
	var oauthSetting model.OAuth2Setting
	if err := h.settingService.GetOAuth2Setting(ctx, &oauthSetting); err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK(ctx, oauthSetting, commonModel.GET_OAUTH_SETTINGS_SUCCESS), nil
}

// UpdateOAuth2Settings 更新 OAuth2 设置。
func (h *SettingHandler) UpdateOAuth2Settings(ctx context.Context, in *OAuth2SettingInput) (*humares.Envelope[any], error) {
	if err := h.settingService.UpdateOAuth2Setting(ctx, &in.Body); err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK[any](ctx, nil, commonModel.UPDATE_OAUTH_SETTINGS_SUCCESS), nil
}

// GetPasskeySettings 获取 Passkey(WebAuthn) 设置。
func (h *SettingHandler) GetPasskeySettings(ctx context.Context, _ *EmptyInput) (*humares.Envelope[model.PasskeySetting], error) {
	var passkeySetting model.PasskeySetting
	if err := h.settingService.GetPasskeySetting(ctx, &passkeySetting); err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK(ctx, passkeySetting, commonModel.GET_PASSKEY_SETTINGS_SUCCESS), nil
}

// UpdatePasskeySettings 更新 Passkey(WebAuthn) 设置。
func (h *SettingHandler) UpdatePasskeySettings(ctx context.Context, in *PasskeySettingInput) (*humares.Envelope[any], error) {
	if err := h.settingService.UpdatePasskeySetting(ctx, &in.Body); err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK[any](ctx, nil, commonModel.UPDATE_PASSKEY_SETTINGS_SUCCESS), nil
}

// GetWebhook 获取所有 Webhook。
func (h *SettingHandler) GetWebhook(ctx context.Context, _ *EmptyInput) (*humares.Envelope[[]webhookModel.Webhook], error) {
	result, err := h.settingService.GetAllWebhooks(ctx)
	if err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK(ctx, result, commonModel.GET_WEBHOOK_SUCCESS), nil
}

// CreateWebhook 创建新的 Webhook。
func (h *SettingHandler) CreateWebhook(ctx context.Context, in *WebhookInput) (*humares.Envelope[any], error) {
	if err := h.settingService.CreateWebhook(ctx, &in.Body); err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK[any](ctx, nil, commonModel.CREATE_WEBHOOK_SUCCESS), nil
}

// UpdateWebhook 根据 ID 更新 Webhook。
func (h *SettingHandler) UpdateWebhook(ctx context.Context, in *WebhookIDBodyInput) (*humares.Envelope[any], error) {
	if err := h.settingService.UpdateWebhook(ctx, in.ID, &in.Body); err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK[any](ctx, nil, commonModel.UPDATE_WEBHOOK_SUCCESS), nil
}

// DeleteWebhook 根据 ID 删除 Webhook。
func (h *SettingHandler) DeleteWebhook(ctx context.Context, in *IDInput) (*humares.Envelope[any], error) {
	if err := h.settingService.DeleteWebhook(ctx, in.ID); err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK[any](ctx, nil, commonModel.DELETE_WEBHOOK_SUCCESS), nil
}

// TestWebhook 根据 ID 触发一次 Webhook 测试请求。
func (h *SettingHandler) TestWebhook(ctx context.Context, in *IDInput) (*humares.Envelope[any], error) {
	if err := h.settingService.TestWebhook(ctx, in.ID); err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK[any](ctx, nil, commonModel.TEST_WEBHOOK_SUCCESS), nil
}

// GetSnapshotScheduleSetting 获取定时快照计划。
func (h *SettingHandler) GetSnapshotScheduleSetting(ctx context.Context, _ *EmptyInput) (*humares.Envelope[model.SnapshotSchedule], error) {
	var snapshotSchedule model.SnapshotSchedule
	if err := h.settingService.GetSnapshotScheduleSetting(&snapshotSchedule); err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK(ctx, snapshotSchedule, commonModel.GET_SETTINGS_SUCCESS), nil
}

// UpdateSnapshotScheduleSetting 设置定时快照计划。
func (h *SettingHandler) UpdateSnapshotScheduleSetting(ctx context.Context, in *SnapshotScheduleInput) (*humares.Envelope[any], error) {
	if err := h.settingService.UpdateSnapshotScheduleSetting(ctx, &in.Body); err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK[any](ctx, nil, commonModel.SCHEDULE_SNAPSHOT_SUCCESS), nil
}

// GetAgentSettings 获取 Agent 设置。
func (h *SettingHandler) GetAgentSettings(ctx context.Context, _ *EmptyInput) (*humares.Envelope[model.AgentSetting], error) {
	var settings model.AgentSetting
	if err := h.settingService.GetAgentSettings(ctx, &settings); err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK(ctx, settings, commonModel.GET_SETTINGS_SUCCESS), nil
}

// UpdateAgentSettings 更新 Agent 设置。
func (h *SettingHandler) UpdateAgentSettings(ctx context.Context, in *AgentSettingInput) (*humares.Envelope[any], error) {
	if err := h.settingService.UpdateAgentSettings(ctx, &in.Body); err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK[any](ctx, nil, commonModel.UPDATE_SETTINGS_SUCCESS), nil
}

// TestAgentConnection 用提交的 Agent 配置做一次最小探活（不保存）。
func (h *SettingHandler) TestAgentConnection(ctx context.Context, in *AgentSettingInput) (*humares.Envelope[any], error) {
	if err := h.settingService.TestAgentConnection(ctx, &in.Body); err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK[any](ctx, nil, commonModel.AGENT_TEST_CONNECTION_SUCCESS), nil
}

// GetEmbeddingSettings 获取 Embedding 向量设置。
func (h *SettingHandler) GetEmbeddingSettings(ctx context.Context, _ *EmptyInput) (*humares.Envelope[model.EmbeddingSetting], error) {
	setting, err := h.settingService.GetEmbeddingSetting(ctx)
	if err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK(ctx, setting, commonModel.GET_SETTINGS_SUCCESS), nil
}

// UpdateEmbeddingSettings 更新 Embedding 向量设置。
func (h *SettingHandler) UpdateEmbeddingSettings(ctx context.Context, in *EmbeddingSettingInput) (*humares.Envelope[any], error) {
	if err := h.settingService.UpdateEmbeddingSetting(ctx, in.Body); err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK[any](ctx, nil, commonModel.UPDATE_SETTINGS_SUCCESS), nil
}

// --- admin:token ---

// ListAccessTokens 列出当前用户的所有访问令牌。
func (h *SettingHandler) ListAccessTokens(ctx context.Context, _ *EmptyInput) (*humares.Envelope[[]model.AccessTokenSetting], error) {
	result, err := h.settingService.ListAccessTokens(ctx)
	if err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK(ctx, result, commonModel.LIST_ACCESS_TOKENS_SUCCESS), nil
}

// CreateAccessToken 为当前用户创建一个新的访问令牌（返回值即明文 token，仅此一次可见）。
func (h *SettingHandler) CreateAccessToken(ctx context.Context, in *AccessTokenInput) (*humares.Envelope[string], error) {
	createdToken, err := h.settingService.CreateAccessToken(ctx, &in.Body)
	if err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK(ctx, createdToken, commonModel.CREATE_ACCESS_TOKEN_SUCCESS), nil
}

// DeleteAccessToken 根据 ID 删除访问令牌。
func (h *SettingHandler) DeleteAccessToken(ctx context.Context, in *IDInput) (*humares.Envelope[any], error) {
	if err := h.settingService.DeleteAccessToken(ctx, in.ID); err != nil {
		return nil, humares.Err(ctx, err)
	}
	return humares.OK[any](ctx, nil, commonModel.DELETE_ACCESS_TOKEN_SUCCESS), nil
}
