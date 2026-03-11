package service

import (
	"context"
	"errors"

	"github.com/lin-snow/ech0/internal/config"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
	httpUtil "github.com/lin-snow/ech0/internal/util/http"
	jsonUtil "github.com/lin-snow/ech0/internal/util/json"
	"github.com/lin-snow/ech0/pkg/viewer"
)

// GetSetting 获取设置
func (settingService *SettingService) GetSetting(setting *model.SystemSetting) error {
	return settingService.transactor.Run(context.Background(), func(ctx context.Context) error {
		systemSetting, err := settingService.keyvalueRepository.GetKeyValue(
			ctx,
			commonModel.SystemSettingsKey,
		)
		if err != nil {
			// 数据库中不存在数据，手动添加初始数据
			setting.SiteTitle = config.Config().Setting.SiteTitle
			setting.ServerLogo = config.Config().Setting.ServerLogo
			setting.ServerName = config.Config().Setting.Servername
			setting.ServerURL = config.Config().Setting.Serverurl
			setting.AllowRegister = config.Config().Setting.AllowRegister
			setting.ICPNumber = config.Config().Setting.Icpnumber
			setting.FooterContent = config.Config().Setting.FooterContent
			setting.FooterLink = config.Config().Setting.FooterLink
			setting.MetingAPI = config.Config().Setting.MetingAPI
			setting.CustomCSS = config.Config().Setting.CustomCSS
			setting.CustomJS = config.Config().Setting.CustomJS

			// 处理 URL
			setting.ServerURL = httpUtil.TrimURL(setting.ServerURL)
			setting.FooterLink = httpUtil.TrimURL(setting.FooterLink)
			setting.MetingAPI = httpUtil.TrimURL(setting.MetingAPI)

			// 序列化为 JSON
			settingToJSON, err := jsonUtil.JSONMarshal(setting)
			if err != nil {
				return err
			}
			if err := settingService.keyvalueRepository.AddKeyValue(ctx, commonModel.SystemSettingsKey, string(settingToJSON)); err != nil {
				return err
			}

			// 处理 ServerURL
			if err := settingService.keyvalueRepository.AddKeyValue(ctx, commonModel.ServerURLKey, setting.ServerURL); err != nil {
				return err
			}

			return nil
		}

		if err := jsonUtil.JSONUnmarshal([]byte(systemSetting), setting); err != nil {
			return err
		}

		return nil
	})
}

// UpdateSetting 更新设置
func (settingService *SettingService) UpdateSetting(
	ctx context.Context,
	newSetting *model.SystemSettingDto,
) error {
	userid := viewer.MustFromContext(ctx).UserID()
	return settingService.transactor.Run(ctx, func(ctx context.Context) error {
		user, err := settingService.commonService.CommonGetUserByUserId(ctx, userid)
		if err != nil {
			return err
		}
		if !user.IsAdmin {
			return errors.New(commonModel.NO_PERMISSION_DENIED)
		}

		var setting model.SystemSetting
		setting.SiteTitle = newSetting.SiteTitle
		setting.ServerLogo = newSetting.ServerLogo
		setting.ServerName = newSetting.ServerName
		setting.ServerURL = httpUtil.TrimURL(newSetting.ServerURL)
		setting.AllowRegister = newSetting.AllowRegister
		setting.ICPNumber = newSetting.ICPNumber
		setting.FooterContent = newSetting.FooterContent
		setting.FooterLink = httpUtil.TrimURL(newSetting.FooterLink)
		setting.MetingAPI = httpUtil.TrimURL(newSetting.MetingAPI)
		setting.CustomCSS = newSetting.CustomCSS
		setting.CustomJS = newSetting.CustomJS

		// 序列化为 JSON
		settingToJSON, err := jsonUtil.JSONMarshal(setting)
		if err != nil {
			return err
		}

		// 将字节切片转换为字符串
		settingToJSONString := string(settingToJSON)
		if err := settingService.keyvalueRepository.UpdateKeyValue(ctx, commonModel.SystemSettingsKey, settingToJSONString); err != nil {
			return err
		}

		// 更新 ServerURL
		if err := settingService.keyvalueRepository.UpdateKeyValue(ctx, commonModel.ServerURLKey, setting.ServerURL); err != nil {
			return err
		}

		return nil
	})
}
