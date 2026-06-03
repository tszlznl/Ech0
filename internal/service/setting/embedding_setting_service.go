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

// GetEmbeddingSetting 获取 Embedding 向量设置。
//
// Embedding 设置随 Chat/RAG 功能引入，归口到 setting 域与其它系统设置统一管理。
// 缺省值（Enable=false）由 setting 引擎处理，与既有 setting 读取语义一致。
func (settingService *SettingService) GetEmbeddingSetting(
	ctx context.Context,
) (model.EmbeddingSetting, error) {
	return coreSetting.Get(ctx, settingService.durableKV, coreSetting.Embedding)
}

// UpdateEmbeddingSetting 更新 Embedding 向量设置。
func (settingService *SettingService) UpdateEmbeddingSetting(
	ctx context.Context,
	dto model.EmbeddingSettingDto,
) error {
	// 鉴权（与其它 Update* 一致；路由已要求 admin scope，这里做服务层 defense-in-depth）。
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := settingService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	setting := model.EmbeddingSetting{
		Enable:    dto.Enable,
		Model:     strings.TrimSpace(dto.Model),
		ApiKey:    strings.TrimSpace(dto.ApiKey),
		BaseURL:   strings.TrimSpace(dto.BaseURL),
		Dim:       dto.Dim,
		BatchSize: dto.BatchSize,
	}

	return coreSetting.Set(ctx, settingService.durableKV, coreSetting.Embedding, setting)
}
