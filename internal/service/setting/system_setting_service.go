package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/lin-snow/ech0/internal/config"
	i18nUtil "github.com/lin-snow/ech0/internal/i18n"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
	httpUtil "github.com/lin-snow/ech0/internal/util/http"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"github.com/lin-snow/ech0/pkg/viewer"
	"go.uber.org/zap"
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
			setting.DefaultLocale = string(commonModel.DefaultLocale)
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
			settingToJSON, err := json.Marshal(setting)
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

		if err := json.Unmarshal([]byte(systemSetting), setting); err != nil {
			return err
		}
		setting.DefaultLocale = i18nUtil.ResolveLocale(setting.DefaultLocale)

		return nil
	})
}

// BootstrapDefaultLocale 在首次部署时把部署者的语言写入站点默认。
// 仅当 system_settings KV 还不存在时落库，避免覆盖站长后续在面板里手动选过的值；
// 入参为空或解析失败时不做任何事，让原有的 zh-CN 兜底逻辑接管。
func (settingService *SettingService) BootstrapDefaultLocale(
	ctx context.Context,
	locale string,
) error {
	resolved := strings.TrimSpace(locale)
	if resolved == "" {
		return nil
	}
	resolved = i18nUtil.ResolveLocale(resolved)
	if resolved == "" || resolved == string(commonModel.DefaultLocale) {
		return nil
	}

	return settingService.transactor.Run(ctx, func(ctx context.Context) error {
		if _, err := settingService.keyvalueRepository.GetKeyValue(ctx, commonModel.SystemSettingsKey); err == nil {
			return nil
		}

		setting := model.SystemSetting{
			SiteTitle:     config.Config().Setting.SiteTitle,
			ServerLogo:    config.Config().Setting.ServerLogo,
			ServerName:    config.Config().Setting.Servername,
			ServerURL:     httpUtil.TrimURL(config.Config().Setting.Serverurl),
			AllowRegister: config.Config().Setting.AllowRegister,
			DefaultLocale: resolved,
			ICPNumber:     config.Config().Setting.Icpnumber,
			FooterContent: config.Config().Setting.FooterContent,
			FooterLink:    httpUtil.TrimURL(config.Config().Setting.FooterLink),
			MetingAPI:     httpUtil.TrimURL(config.Config().Setting.MetingAPI),
			CustomCSS:     config.Config().Setting.CustomCSS,
			CustomJS:      config.Config().Setting.CustomJS,
		}
		settingToJSON, err := json.Marshal(setting)
		if err != nil {
			return err
		}
		if err := settingService.keyvalueRepository.AddKeyValue(ctx, commonModel.SystemSettingsKey, string(settingToJSON)); err != nil {
			return err
		}
		return settingService.keyvalueRepository.AddKeyValue(ctx, commonModel.ServerURLKey, setting.ServerURL)
	})
}

// UpdateSetting 更新设置
func (settingService *SettingService) UpdateSetting(
	ctx context.Context,
	newSetting *model.SystemSettingDto,
) error {
	userid := viewer.MustFromContext(ctx).UserID()
	serverLogoChanged := false
	if newSetting != nil {
		var current model.SystemSetting
		if err := settingService.GetSetting(&current); err == nil {
			serverLogoChanged = strings.TrimSpace(current.ServerLogo) != strings.TrimSpace(newSetting.ServerLogo)
		}
	}
	if err := settingService.transactor.Run(ctx, func(ctx context.Context) error {
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
		setting.DefaultLocale = i18nUtil.ResolveLocale(newSetting.DefaultLocale)
		setting.ICPNumber = newSetting.ICPNumber
		setting.FooterContent = newSetting.FooterContent
		setting.FooterLink = httpUtil.TrimURL(newSetting.FooterLink)
		setting.MetingAPI = httpUtil.TrimURL(newSetting.MetingAPI)
		setting.CustomCSS = newSetting.CustomCSS
		setting.CustomJS = newSetting.CustomJS

		// 序列化为 JSON
		settingToJSON, err := json.Marshal(setting)
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
	}); err != nil {
		return err
	}
	if serverLogoChanged && strings.TrimSpace(newSetting.ServerLogoFileID) != "" {
		if err := settingService.fileService.ConfirmTempFiles(ctx, []string{newSetting.ServerLogoFileID}); err != nil {
			logUtil.GetLogger().Warn("confirm temp server logo file failed", zap.Error(err))
		}
	}
	return nil
}
