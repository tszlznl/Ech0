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

// GetAgentInfo 获取 Agent 信息（公开读，缺省值由 setting 引擎处理）。
func (settingService *SettingService) GetAgentInfo(setting *model.AgentSetting) error {
	v, err := coreSetting.Get(context.Background(), settingService.durableKV, coreSetting.Agent)
	if err != nil {
		return err
	}
	*setting = v
	return nil
}

// GetAgentSettings 获取 Agent 设置（管理员可见全量）。
func (settingService *SettingService) GetAgentSettings(
	ctx context.Context,
	setting *model.AgentSetting,
) error {
	// 检查用户权限
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := settingService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	v, err := coreSetting.Get(ctx, settingService.durableKV, coreSetting.Agent)
	if err != nil {
		return err
	}
	*setting = v
	return nil
}

// UpdateAgentSettings 更新 Agent 设置
func (settingService *SettingService) UpdateAgentSettings(
	ctx context.Context,
	newSetting *model.AgentSettingDto,
) error {
	// 检查用户权限
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := settingService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	if newSetting.Protocol != string(commonModel.OpenAI) &&
		newSetting.Protocol != string(commonModel.Anthropic) {
		// 未识别的接口协议（含已下线的 gemini）一律按 OpenAI 兼容协议处理
		newSetting.Protocol = string(commonModel.OpenAI)
	}

	setting := model.AgentSetting{
		Enable:     newSetting.Enable,
		Protocol:   newSetting.Protocol,
		Model:      newSetting.Model,
		ApiKey:     newSetting.ApiKey,
		Prompt:     newSetting.Prompt,
		BaseURL:    urlUtil.TrimURL(newSetting.BaseURL),
		Multimodal: newSetting.Multimodal,
		// 负数视为未配置，归零走保守默认。
		ContextWindow: max(0, newSetting.ContextWindow),
	}
	return coreSetting.Set(ctx, settingService.durableKV, coreSetting.Agent, setting)
}
