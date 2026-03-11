package service

import (
	"context"
	"errors"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
	httpUtil "github.com/lin-snow/ech0/internal/util/http"
	jsonUtil "github.com/lin-snow/ech0/internal/util/json"
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

			// 序列化为 JSON
			settingToJSON, err := jsonUtil.JSONMarshal(setting)
			if err != nil {
				return err
			}
			if err := settingService.keyvalueRepository.AddKeyValue(ctx, commonModel.OAuth2SettingKey, string(settingToJSON)); err != nil {
				return err
			}

			return nil
		}

		if err := jsonUtil.JSONUnmarshal([]byte(oauthSetting), setting); err != nil {
			return err
		}

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
			Enable:       newSetting.Enable,
			Provider:     newSetting.Provider,
			ClientID:     newSetting.ClientID,
			ClientSecret: newSetting.ClientSecret,
			AuthURL:      httpUtil.TrimURL(newSetting.AuthURL),
			TokenURL:     httpUtil.TrimURL(newSetting.TokenURL),
			UserInfoURL:  httpUtil.TrimURL(newSetting.UserInfoURL),
			RedirectURI:  httpUtil.TrimURL(newSetting.RedirectURI),
			Scopes:       newSetting.Scopes,
			IsOIDC:       newSetting.IsOIDC,
			Issuer:       newSetting.Issuer,
			JWKSURL:      httpUtil.TrimURL(newSetting.JWKSURL),
		}

		// 序列化为 JSON
		settingToJSON, err := jsonUtil.JSONMarshal(oauthSetting)
		if err != nil {
			return err
		}

		if err := settingService.keyvalueRepository.UpdateKeyValue(ctx, commonModel.OAuth2SettingKey, string(settingToJSON)); err != nil {
			return err
		}

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

	return nil
}
