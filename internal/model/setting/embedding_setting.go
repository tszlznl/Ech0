// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

// EmbeddingSetting 定义向量 Embedding 设置实体（独立于 Agent 的生成 LLM 配置）。
//
// v1 仅支持 OpenAI 兼容的 /v1/embeddings 接口（覆盖 OpenAI、Qwen/DashScope、
// Ollama、Jina 等绝大多数提供商），因此不带 Protocol 字段。
type EmbeddingSetting struct {
	Enable    bool   `json:"enable"`     // 是否启用 Embedding（Chat 检索的前置条件）
	Model     string `json:"model"`      // Embedding 模型名，如 text-embedding-3-small
	ApiKey    string `json:"api_key"`    // API Key（本地服务如 Ollama 可留空）
	BaseURL   string `json:"base_url"`   // 自定义 API URL（可选）
	Dim       int    `json:"dim"`        // 向量维度，必须与所选模型一致（vec0 建表时写死）
	BatchSize int    `json:"batch_size"` // 单次请求最多向量化的文本条数（0=用默认值）；部分提供商限制 64/25 条
}

// EmbeddingSettingDto 是更新 Embedding 设置的入参
type EmbeddingSettingDto struct {
	Enable    bool   `json:"enable"`
	Model     string `json:"model"`
	ApiKey    string `json:"api_key"`
	BaseURL   string `json:"base_url"`
	Dim       int    `json:"dim"`
	BatchSize int    `json:"batch_size"`
}
