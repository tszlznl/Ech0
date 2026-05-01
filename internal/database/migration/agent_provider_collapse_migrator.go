// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package migration

import (
	"encoding/json"
	"errors"
	"fmt"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	"gorm.io/gorm"
)

// agentProviderCollapseMigrator 把已废弃的 Agent provider 取值
// (deepseek/qwen/ollama/custom) 收敛为新的 3 个枚举值之一。
// 收敛规则：除 anthropic / gemini 外，其余取值统一映射到 openai
// （新枚举里 openai 同时承担 OpenAI 兼容协议下的所有第三方服务）。
type agentProviderCollapseMigrator struct{}

func NewAgentProviderCollapseMigrator() Migrator {
	return &agentProviderCollapseMigrator{}
}

func (m *agentProviderCollapseMigrator) Name() string {
	return "agent_provider_collapse_migrator"
}

func (m *agentProviderCollapseMigrator) Key() string {
	return commonModel.AgentProviderCollapsedKey
}

func (m *agentProviderCollapseMigrator) CanRerun() bool {
	return false
}

func (m *agentProviderCollapseMigrator) Migrate(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	var kv commonModel.KeyValue
	err := db.Where("key = ?", commonModel.AgentSettingKey).First(&kv).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 未配置 Agent，无需迁移
		return nil
	}
	if err != nil {
		return err
	}

	// 用宽松结构解析，避免与具体业务模型耦合
	raw := map[string]any{}
	if err := json.Unmarshal([]byte(kv.Value), &raw); err != nil {
		return fmt.Errorf("agent setting json invalid: %w", err)
	}

	provider, _ := raw["provider"].(string)
	mapped := collapseAgentProvider(provider)
	if mapped == provider {
		return nil
	}

	raw["provider"] = mapped
	encoded, err := json.Marshal(raw)
	if err != nil {
		return err
	}

	kv.Value = string(encoded)
	return db.Save(&kv).Error
}

func collapseAgentProvider(old string) string {
	switch old {
	case string(commonModel.Anthropic), string(commonModel.Gemini), string(commonModel.OpenAI):
		return old
	default:
		// deepseek / qwen / ollama / custom / 空 / 未知 → openai
		return string(commonModel.OpenAI)
	}
}
