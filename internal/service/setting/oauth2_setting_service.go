// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/lin-snow/ech0/internal/config"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
	httpUtil "github.com/lin-snow/ech0/internal/util/http"
	"github.com/lin-snow/ech0/pkg/viewer"
)

// GetOAuth2Setting 获取 OAuth2 设置
func (settingService *SettingService) GetOAuth2Setting(
	ctx context.Context,
	setting *model.OAuth2Setting,
	forInternal bool,
) error {
	userid := viewer.MustFromContext(ctx).UserID()
	return settingService.transactor.Run(ctx, func(ctx context.Context) error {
		if !forInternal {
			user, err := settingService.commonService.CommonGetUserByUserId(ctx, userid)
			if err != nil {
				return err
			}
			if !user.IsAdmin {
				return errors.New(commonModel.NO_PERMISSION_DENIED)
			}
		}

		oauthSetting, err := settingService.keyvalueRepository.GetKeyValue(
			ctx,
			commonModel.OAuth2SettingKey,
		)
		if err != nil {
			// 数据库中不存在数据，手动添加初始数据
			setting.Enable = false
			setting.Provider = string(commonModel.OAuth2GITHUB)
			setting.ClientID = ""
			setting.ClientSecret = ""
			setting.AuthURL = "https://github.com/login/oauth/authorize"
			setting.TokenURL = "https://github.com/login/oauth/access_token"
			setting.UserInfoURL = "https://api.github.com/user"
			setting.RedirectURI = ""
			setting.Scopes = []string{
				"read:user",
			}
			setting.IsOIDC = false
			setting.Issuer = ""
			setting.JWKSURL = ""
			setting.AuthRedirectAllowedReturnURLs = append([]string{}, config.Config().Auth.Redirect.AllowedReturnURLs...)
			setting.CORSAllowedOrigins = append([]string{}, config.Config().Web.CORS.AllowedOrigins...)

			// 序列化为 JSON
			settingToJSON, err := json.Marshal(setting)
			if err != nil {
				return err
			}
			if err := settingService.keyvalueRepository.AddKeyValue(ctx, commonModel.OAuth2SettingKey, string(settingToJSON)); err != nil {
				return err
			}

			return nil
		}

		if err := json.Unmarshal([]byte(oauthSetting), setting); err != nil {
			return err
		}
		applyOAuthBoundaryFallback(setting)
		syncRuntimeOAuthBoundary(setting)

		return nil
	})
}

// UpdateOAuth2Setting 更新 OAuth2 设置
func (settingService *SettingService) UpdateOAuth2Setting(
	ctx context.Context,
	newSetting *model.OAuth2SettingDto,
) error {
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := settingService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	return settingService.transactor.Run(ctx, func(ctx context.Context) error {
		oauthSetting := &model.OAuth2Setting{
			Enable:                        newSetting.Enable,
			Provider:                      newSetting.Provider,
			ClientID:                      newSetting.ClientID,
			ClientSecret:                  newSetting.ClientSecret,
			AuthURL:                       httpUtil.TrimURL(newSetting.AuthURL),
			TokenURL:                      httpUtil.TrimURL(newSetting.TokenURL),
			UserInfoURL:                   httpUtil.TrimURL(newSetting.UserInfoURL),
			RedirectURI:                   httpUtil.TrimURL(newSetting.RedirectURI),
			Scopes:                        newSetting.Scopes,
			IsOIDC:                        newSetting.IsOIDC,
			Issuer:                        newSetting.Issuer,
			JWKSURL:                       httpUtil.TrimURL(newSetting.JWKSURL),
			AuthRedirectAllowedReturnURLs: sanitizeURLList(newSetting.AuthRedirectAllowedReturnURLs),
			CORSAllowedOrigins:            sanitizeURLList(newSetting.CORSAllowedOrigins),
		}
		applyOAuthBoundaryFallback(oauthSetting)

		// 序列化为 JSON
		settingToJSON, err := json.Marshal(oauthSetting)
		if err != nil {
			return err
		}

		if err := settingService.keyvalueRepository.UpdateKeyValue(ctx, commonModel.OAuth2SettingKey, string(settingToJSON)); err != nil {
			return err
		}
		// 同步运行时配置，确保中间件/服务按 Panel 最新配置生效。
		syncRuntimeOAuthBoundary(oauthSetting)

		return nil
	})
}

// GetOAuth2Status 获取 OAuth2 状态
func (settingService *SettingService) GetOAuth2Status(status *model.OAuth2Status) error {
	var oauthSetting model.OAuth2Setting
	systemCtx := viewer.WithContext(context.Background(), viewer.NewSystemViewer())
	if err := settingService.GetOAuth2Setting(systemCtx, &oauthSetting, true); err != nil {
		return err
	}

	status.Enabled = oauthSetting.Enable
	status.Provider = oauthSetting.Provider
	status.OAuthReady = len(oauthSetting.AuthRedirectAllowedReturnURLs) > 0 && len(oauthSetting.CORSAllowedOrigins) > 0

	return nil
}

func sanitizeURLList(values []string) []string {
	result := make([]string, 0, len(values))
	for _, v := range values {
		if trimmed := httpUtil.TrimURL(v); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func applyOAuthBoundaryFallback(setting *model.OAuth2Setting) {
	if len(setting.AuthRedirectAllowedReturnURLs) == 0 {
		setting.AuthRedirectAllowedReturnURLs = append(
			[]string{},
			config.Config().Auth.Redirect.AllowedReturnURLs...,
		)
	}
	if len(setting.CORSAllowedOrigins) == 0 {
		setting.CORSAllowedOrigins = append([]string{}, config.Config().Web.CORS.AllowedOrigins...)
	}
}

func syncRuntimeOAuthBoundary(setting *model.OAuth2Setting) {
	cfg := config.Config()
	cfg.Auth.Redirect.AllowedReturnURLs = append([]string{}, setting.AuthRedirectAllowedReturnURLs...)
	cfg.Web.CORS.AllowedOrigins = append([]string{}, setting.CORSAllowedOrigins...)
}
