// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package embedding 封装向量 Embedding 的外部 API 调用（OpenAI 兼容 /v1/embeddings）。
// 与 internal/agent 平级：agent 负责文本生成，embedding 负责向量化。
package embedding

import (
	"context"
	"errors"

	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	openai "github.com/sashabaranov/go-openai"
)

var (
	// ErrNotEnabled 表示 Embedding 功能未启用
	ErrNotEnabled = errors.New("embedding: not enabled")
	// ErrModelMissing 表示未配置 Embedding 模型
	ErrModelMissing = errors.New("embedding: model missing")
	// ErrEmptyResponse 表示服务端返回空结果
	ErrEmptyResponse = errors.New("embedding: empty response")
)

// Embed 批量生成文本向量（OpenAI 兼容 /v1/embeddings）。
// 返回的切片顺序与 inputs 一一对应。
func Embed(
	ctx context.Context,
	setting settingModel.EmbeddingSetting,
	inputs []string,
) ([][]float32, error) {
	if !setting.Enable {
		return nil, ErrNotEnabled
	}
	if setting.Model == "" {
		return nil, ErrModelMissing
	}
	if len(inputs) == 0 {
		return nil, nil
	}

	cfg := openai.DefaultConfig(setting.ApiKey)
	if setting.BaseURL != "" {
		cfg.BaseURL = setting.BaseURL
	}
	client := openai.NewClientWithConfig(cfg)

	resp, err := client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Model: openai.EmbeddingModel(setting.Model),
		Input: inputs,
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Data) != len(inputs) {
		return nil, ErrEmptyResponse
	}

	out := make([][]float32, len(resp.Data))
	for i := range resp.Data {
		out[i] = resp.Data[i].Embedding
	}
	return out, nil
}

// EmbedOne 生成单条文本向量。
func EmbedOne(
	ctx context.Context,
	setting settingModel.EmbeddingSetting,
	input string,
) ([]float32, error) {
	vecs, err := Embed(ctx, setting, []string{input})
	if err != nil {
		return nil, err
	}
	if len(vecs) == 0 {
		return nil, ErrEmptyResponse
	}
	return vecs[0], nil
}
