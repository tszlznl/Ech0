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

// agentSettingProtocolRenameMigrator 把存量 agent_setting JSON 里的
// "provider" 字段重命名为 "protocol"，与代码层把 Agent "提供商" 概念
// 统一为 "接口协议" 保持一致。
//
// 必须在 agentProtocolCollapseMigrator 之后执行：collapse 先在旧字段名
// "provider" 上完成枚举收敛，本迁移再把字段名整体改为 "protocol"。
// 不做此迁移，老用户已选的 anthropic/gemini 会在反序列化时丢失，被强制降级回 openai。
type agentSettingProtocolRenameMigrator struct{}

func NewAgentSettingProtocolRenameMigrator() Migrator {
	return &agentSettingProtocolRenameMigrator{}
}

func (m *agentSettingProtocolRenameMigrator) Name() string {
	return "agent_setting_protocol_rename_migrator"
}

func (m *agentSettingProtocolRenameMigrator) Key() string {
	return commonModel.AgentSettingProtocolRenamedKey
}

func (m *agentSettingProtocolRenameMigrator) CanRerun() bool {
	return false
}

func (m *agentSettingProtocolRenameMigrator) Migrate(db *gorm.DB) error {
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

	// 已经是新字段名，幂等返回
	if _, ok := raw["provider"]; !ok {
		return nil
	}

	// 仅当尚未存在 protocol 时才以 provider 的值填充，避免覆盖新值
	if _, ok := raw["protocol"]; !ok {
		raw["protocol"] = raw["provider"]
	}
	delete(raw, "provider")

	encoded, err := json.Marshal(raw)
	if err != nil {
		return err
	}

	kv.Value = string(encoded)
	return db.Save(&kv).Error
}
