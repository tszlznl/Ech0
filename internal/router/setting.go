// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/handler/humares"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	authService "github.com/lin-snow/ech0/internal/service/auth"
)

// registerSettingHuma 注册系统设置路由（全部 JSON）。
func registerSettingHuma(api huma.API, h *handler.Bundle, revoker authService.TokenRevoker) {
	pub := func(id, method, path, summary string) huma.Operation {
		return huma.Operation{OperationID: id, Method: method, Path: path, Summary: summary, Tags: []string{"Setting"}}
	}
	adminSettings := humares.Secured(authModel.ScopeAdminSettings)
	adminSettingsMW := securedMW(revoker, authModel.ScopeAdminSettings)
	admin := func(id, method, path, summary string) huma.Operation {
		return huma.Operation{OperationID: id, Method: method, Path: path, Summary: summary, Tags: []string{"Setting"}, Security: adminSettings, Middlewares: adminSettingsMW}
	}
	adminToken := humares.Secured(authModel.ScopeAdminToken)
	adminTokenMW := securedMW(revoker, authModel.ScopeAdminToken)
	token := func(id, method, path, summary string) huma.Operation {
		return huma.Operation{OperationID: id, Method: method, Path: path, Summary: summary, Tags: []string{"Setting"}, Security: adminToken, Middlewares: adminTokenMW}
	}

	// 公开
	reg(api, pub("settings-get", http.MethodGet, "/settings", "获取系统全局设置"), h.SettingHandler.GetSettings)
	reg(api, pub("oauth2-status", http.MethodGet, "/oauth2/status", "获取 OAuth2 状态"), h.SettingHandler.GetOAuth2Status)
	reg(api, pub("passkey-status", http.MethodGet, "/passkey/status", "获取 Passkey 状态"), h.SettingHandler.GetPasskeyStatus)
	reg(api, pub("agent-info", http.MethodGet, "/agent/info", "获取 Agent 公开信息"), h.SettingHandler.GetAgentInfo)

	// admin:settings
	reg(api, admin("settings-update", http.MethodPut, "/settings", "更新系统全局设置"), h.SettingHandler.UpdateSettings)
	reg(api, admin("s3-get", http.MethodGet, "/s3/settings", "获取 S3 存储设置"), h.SettingHandler.GetS3Settings)
	reg(api, admin("s3-update", http.MethodPut, "/s3/settings", "更新 S3 存储设置"), h.SettingHandler.UpdateS3Settings)
	reg(api, admin("s3-test", http.MethodPost, "/s3/settings/test", "测试 S3 存储连接"), h.SettingHandler.TestS3Connection)
	reg(api, admin("oauth2-get", http.MethodGet, "/oauth2/settings", "获取 OAuth2 设置"), h.SettingHandler.GetOAuth2Settings)
	reg(api, admin("oauth2-update", http.MethodPut, "/oauth2/settings", "更新 OAuth2 设置"), h.SettingHandler.UpdateOAuth2Settings)
	reg(api, admin("passkey-get", http.MethodGet, "/passkey/settings", "获取 Passkey 设置"), h.SettingHandler.GetPasskeySettings)
	reg(api, admin("passkey-update", http.MethodPut, "/passkey/settings", "更新 Passkey 设置"), h.SettingHandler.UpdatePasskeySettings)
	reg(api, admin("webhook-list", http.MethodGet, "/webhook", "获取所有 Webhook"), h.SettingHandler.GetWebhook)
	reg(api, admin("webhook-create", http.MethodPost, "/webhook", "创建 Webhook"), h.SettingHandler.CreateWebhook)
	reg(api, admin("webhook-update", http.MethodPut, "/webhook/{id}", "更新 Webhook"), h.SettingHandler.UpdateWebhook)
	reg(api, admin("webhook-delete", http.MethodDelete, "/webhook/{id}", "删除 Webhook"), h.SettingHandler.DeleteWebhook)
	reg(api, admin("webhook-test", http.MethodPost, "/webhook/{id}/test", "测试 Webhook"), h.SettingHandler.TestWebhook)
	reg(api, admin("snapshot-schedule-get", http.MethodGet, "/snapshot/schedule", "获取定时快照计划"), h.SettingHandler.GetSnapshotScheduleSetting)
	reg(api, admin("snapshot-schedule-update", http.MethodPost, "/snapshot/schedule", "设置定时快照计划"), h.SettingHandler.UpdateSnapshotScheduleSetting)
	reg(api, admin("agent-settings-get", http.MethodGet, "/agent/settings", "获取 Agent 设置"), h.SettingHandler.GetAgentSettings)
	reg(api, admin("agent-settings-update", http.MethodPut, "/agent/settings", "更新 Agent 设置"), h.SettingHandler.UpdateAgentSettings)
	reg(api, admin("agent-settings-test", http.MethodPost, "/agent/settings/test", "测试 Copilot 连接"), h.SettingHandler.TestAgentConnection)
	reg(api, admin("embedding-settings-get", http.MethodGet, "/embedding/settings", "获取 Embedding 设置"), h.SettingHandler.GetEmbeddingSettings)
	reg(api, admin("embedding-settings-update", http.MethodPut, "/embedding/settings", "更新 Embedding 设置"), h.SettingHandler.UpdateEmbeddingSettings)

	// admin:token
	reg(api, token("access-token-list", http.MethodGet, "/access-tokens", "列出访问令牌"), h.SettingHandler.ListAccessTokens)
	reg(api, token("access-token-create", http.MethodPost, "/access-tokens", "创建访问令牌"), h.SettingHandler.CreateAccessToken)
	reg(api, token("access-token-delete", http.MethodDelete, "/access-tokens/{id}", "删除访问令牌"), h.SettingHandler.DeleteAccessToken)
}
