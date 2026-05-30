// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"encoding/json"
	"strings"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
)

// GetEmbeddingSetting 获取 Embedding 向量设置。
//
// Embedding 设置随 Chat/RAG 功能引入，归口到 setting 域与其它系统设置统一管理。
// 数据库中不存在配置时返回零值（Enable=false），与既有 setting 读取语义一致。
func (settingService *SettingService) GetEmbeddingSetting(
	ctx context.Context,
) (model.EmbeddingSetting, error) {
	var setting model.EmbeddingSetting

	raw, err := settingService.keyvalueRepository.GetKeyValue(ctx, commonModel.EmbeddingSettingKey)
	if err != nil {
		// 未配置 → 返回零值（Enable=false）
		return setting, nil
	}
	if err := json.Unmarshal([]byte(raw), &setting); err != nil {
		return setting, err
	}
	return setting, nil
}

// UpdateEmbeddingSetting 更新 Embedding 向量设置。
func (settingService *SettingService) UpdateEmbeddingSetting(
	ctx context.Context,
	dto model.EmbeddingSettingDto,
) error {
	setting := model.EmbeddingSetting{
		Enable:  dto.Enable,
		Model:   strings.TrimSpace(dto.Model),
		ApiKey:  strings.TrimSpace(dto.ApiKey),
		BaseURL: strings.TrimSpace(dto.BaseURL),
		Dim:     dto.Dim,
	}

	encoded, err := json.Marshal(setting)
	if err != nil {
		return err
	}
	return settingService.keyvalueRepository.AddOrUpdateKeyValue(
		ctx,
		commonModel.EmbeddingSettingKey,
		string(encoded),
	)
}
