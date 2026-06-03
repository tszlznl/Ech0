// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"errors"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
	coreSetting "github.com/lin-snow/ech0/internal/setting"
	urlUtil "github.com/lin-snow/ech0/internal/util/url"
	"github.com/lin-snow/ech0/pkg/viewer"
)

// GetOAuth2Setting 获取 OAuth2 设置（管理员可见全量；缺省/归一化由 setting 引擎处理）。
// 内部调用方（auth 等）直接走 setting.Get(ctx, kv, setting.OAuth2)，不经此鉴权方法。
func (settingService *SettingService) GetOAuth2Setting(
	ctx context.Context,
	setting *model.OAuth2Setting,
) error {
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := settingService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	v, err := coreSetting.Get(ctx, settingService.durableKV, coreSetting.OAuth2)
	if err != nil {
		return err
	}
	*setting = v
	return nil
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

	oauthSetting := model.OAuth2Setting{
		Enable:                        newSetting.Enable,
		Provider:                      newSetting.Provider,
		ClientID:                      newSetting.ClientID,
		ClientSecret:                  newSetting.ClientSecret,
		AuthURL:                       urlUtil.TrimURL(newSetting.AuthURL),
		TokenURL:                      urlUtil.TrimURL(newSetting.TokenURL),
		UserInfoURL:                   urlUtil.TrimURL(newSetting.UserInfoURL),
		RedirectURI:                   urlUtil.TrimURL(newSetting.RedirectURI),
		Scopes:                        newSetting.Scopes,
		IsOIDC:                        newSetting.IsOIDC,
		Issuer:                        newSetting.Issuer,
		JWKSURL:                       urlUtil.TrimURL(newSetting.JWKSURL),
		AuthRedirectAllowedReturnURLs: sanitizeURLList(newSetting.AuthRedirectAllowedReturnURLs),
		CORSAllowedOrigins:            sanitizeURLList(newSetting.CORSAllowedOrigins),
	}
	// 边界白名单为空时回退到 config 默认，由 coreSetting.Set 的 Normalize 统一处理。
	return coreSetting.Set(ctx, settingService.durableKV, coreSetting.OAuth2, oauthSetting)
}

// GetOAuth2Status 获取 OAuth2 状态（公开读，直接走 setting 引擎）。
func (settingService *SettingService) GetOAuth2Status(status *model.OAuth2Status) error {
	oauthSetting, err := coreSetting.Get(context.Background(), settingService.durableKV, coreSetting.OAuth2)
	if err != nil {
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
		if trimmed := urlUtil.TrimURL(v); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
