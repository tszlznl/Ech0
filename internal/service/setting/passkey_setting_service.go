package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/lin-snow/ech0/internal/config"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
	"github.com/lin-snow/ech0/pkg/viewer"
)

// GetPasskeySetting 获取 Passkey 设置
func (settingService *SettingService) GetPasskeySetting(
	ctx context.Context,
	setting *model.PasskeySetting,
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

		passkeySetting, err := settingService.keyvalueRepository.GetKeyValue(
			ctx,
			commonModel.PasskeySettingKey,
		)
		if err != nil {
			// 首次读取时优先从旧 oauth2_setting 中迁移 WebAuthn 字段，避免升级后配置丢失。
			if migrated, ok := settingService.readLegacyPasskeySetting(ctx); ok {
				*setting = migrated
			} else {
				setting.WebAuthnRPID = strings.TrimSpace(config.Config().Auth.WebAuthn.RPID)
				setting.WebAuthnAllowedOrigins = append([]string{}, config.Config().Auth.WebAuthn.Origins...)
			}
			applyPasskeyBoundaryFallback(setting)
			settingToJSON, marshalErr := json.Marshal(setting)
			if marshalErr != nil {
				return marshalErr
			}
			if addErr := settingService.keyvalueRepository.AddKeyValue(ctx, commonModel.PasskeySettingKey, string(settingToJSON)); addErr != nil {
				return addErr
			}
			syncRuntimePasskeyBoundary(setting)
			return nil
		}

		if err := json.Unmarshal([]byte(passkeySetting), setting); err != nil {
			return err
		}
		applyPasskeyBoundaryFallback(setting)
		syncRuntimePasskeyBoundary(setting)
		return nil
	})
}

// UpdatePasskeySetting 更新 Passkey 设置
func (settingService *SettingService) UpdatePasskeySetting(
	ctx context.Context,
	newSetting *model.PasskeySettingDto,
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
		passkeySetting := &model.PasskeySetting{
			WebAuthnRPID:           strings.TrimSpace(newSetting.WebAuthnRPID),
			WebAuthnAllowedOrigins: sanitizeURLList(newSetting.WebAuthnAllowedOrigins),
		}
		applyPasskeyBoundaryFallback(passkeySetting)

		settingToJSON, err := json.Marshal(passkeySetting)
		if err != nil {
			return err
		}
		if err := settingService.keyvalueRepository.AddOrUpdateKeyValue(
			ctx,
			commonModel.PasskeySettingKey,
			string(settingToJSON),
		); err != nil {
			return err
		}
		syncRuntimePasskeyBoundary(passkeySetting)
		return nil
	})
}

// GetPasskeyStatus 获取 Passkey 状态
func (settingService *SettingService) GetPasskeyStatus(status *model.PasskeyStatus) error {
	var passkeySetting model.PasskeySetting
	systemCtx := viewer.WithContext(context.Background(), viewer.NewSystemViewer())
	if err := settingService.GetPasskeySetting(systemCtx, &passkeySetting, true); err != nil {
		return err
	}
	status.PasskeyReady = strings.TrimSpace(passkeySetting.WebAuthnRPID) != "" &&
		len(passkeySetting.WebAuthnAllowedOrigins) > 0
	return nil
}

func applyPasskeyBoundaryFallback(setting *model.PasskeySetting) {
	if strings.TrimSpace(setting.WebAuthnRPID) == "" {
		setting.WebAuthnRPID = strings.TrimSpace(config.Config().Auth.WebAuthn.RPID)
	}
	if len(setting.WebAuthnAllowedOrigins) == 0 {
		setting.WebAuthnAllowedOrigins = append([]string{}, config.Config().Auth.WebAuthn.Origins...)
	}
}

func syncRuntimePasskeyBoundary(setting *model.PasskeySetting) {
	cfg := config.Config()
	cfg.Auth.WebAuthn.RPID = strings.TrimSpace(setting.WebAuthnRPID)
	cfg.Auth.WebAuthn.Origins = append([]string{}, setting.WebAuthnAllowedOrigins...)
}

func (settingService *SettingService) readLegacyPasskeySetting(ctx context.Context) (model.PasskeySetting, bool) {
	type legacyOAuth2Boundary struct {
		WebAuthnRPID           string   `json:"webauthn_rp_id"`
		WebAuthnAllowedOrigins []string `json:"webauthn_allowed_origins"`
	}
	var result model.PasskeySetting
	raw, err := settingService.keyvalueRepository.GetKeyValue(ctx, commonModel.OAuth2SettingKey)
	if err != nil || strings.TrimSpace(raw) == "" {
		return result, false
	}
	var legacy legacyOAuth2Boundary
	if err := json.Unmarshal([]byte(raw), &legacy); err != nil {
		return result, false
	}
	result.WebAuthnRPID = strings.TrimSpace(legacy.WebAuthnRPID)
	result.WebAuthnAllowedOrigins = sanitizeURLList(legacy.WebAuthnAllowedOrigins)
	return result, result.WebAuthnRPID != "" || len(result.WebAuthnAllowedOrigins) > 0
}
