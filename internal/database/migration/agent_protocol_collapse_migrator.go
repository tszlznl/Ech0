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

// agentProtocolCollapseMigrator 把已废弃的 Agent 协议取值
// (deepseek/qwen/ollama/custom) 收敛为新的 3 个枚举值之一。
// 收敛规则：除 anthropic / gemini 外，其余取值统一映射到 openai
// （新枚举里 openai 同时承担 OpenAI 兼容协议下的所有第三方服务）。
//
// 注意：此迁移作用于历史数据，存量 JSON 仍使用 "provider" 字段名，
// 故这里读写的仍是 raw["provider"]；字段名到 "protocol" 的重命名由
// agentSettingProtocolRenameMigrator 在其之后完成。
type agentProtocolCollapseMigrator struct{}

func NewAgentProtocolCollapseMigrator() Migrator {
	return &agentProtocolCollapseMigrator{}
}

func (m *agentProtocolCollapseMigrator) Name() string {
	return "agent_provider_collapse_migrator"
}

func (m *agentProtocolCollapseMigrator) Key() string {
	return commonModel.AgentProtocolCollapsedKey
}

func (m *agentProtocolCollapseMigrator) CanRerun() bool {
	return false
}

func (m *agentProtocolCollapseMigrator) Migrate(db *gorm.DB) error {
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
	mapped := collapseAgentProtocol(provider)
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

func collapseAgentProtocol(old string) string {
	switch old {
	// gemini 已整协议下线，但此处保留其历史直通映射，忠实还原迁移当时的行为；
	// 运行时 gemini 会落到 AGENT_PROTOCOL_NOT_FOUND（见 internal/agent）。
	case string(commonModel.Anthropic), "gemini", string(commonModel.OpenAI):
		return old
	default:
		// deepseek / qwen / ollama / custom / 空 / 未知 → openai
		return string(commonModel.OpenAI)
	}
}
