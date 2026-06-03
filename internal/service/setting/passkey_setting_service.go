// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"errors"
	"strings"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
	coreSetting "github.com/lin-snow/ech0/internal/setting"
	"github.com/lin-snow/ech0/pkg/viewer"
)

// GetPasskeySetting 获取 Passkey 设置（管理员可见全量；缺省/归一化由 setting 引擎处理，
// 旧 oauth2_setting 的 WebAuthn 字段迁移已在启动 seeder 的 Passkey.Migrate 中完成）。
func (settingService *SettingService) GetPasskeySetting(
	ctx context.Context,
	setting *model.PasskeySetting,
) error {
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := settingService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	v, err := coreSetting.Get(ctx, settingService.durableKV, coreSetting.Passkey)
	if err != nil {
		return err
	}
	*setting = v
	return nil
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

	passkeySetting := model.PasskeySetting{
		WebAuthnRPID:           strings.TrimSpace(newSetting.WebAuthnRPID),
		WebAuthnAllowedOrigins: sanitizeURLList(newSetting.WebAuthnAllowedOrigins),
	}
	// RPID/Origins 为空时回退到 config 默认，由 coreSetting.Set 的 Normalize 统一处理。
	return coreSetting.Set(ctx, settingService.durableKV, coreSetting.Passkey, passkeySetting)
}

// GetPasskeyStatus 获取 Passkey 状态（公开读，直接走 setting 引擎）。
func (settingService *SettingService) GetPasskeyStatus(status *model.PasskeyStatus) error {
	v, err := coreSetting.Get(context.Background(), settingService.durableKV, coreSetting.Passkey)
	if err != nil {
		return err
	}
	status.PasskeyReady = strings.TrimSpace(v.WebAuthnRPID) != "" &&
		len(v.WebAuthnAllowedOrigins) > 0
	return nil
}
