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

// defaultBatchSize 是未配置批次大小时，单次 /v1/embeddings 请求的文本条数上限。
// 取 64 是因为不少国产提供商（如 Qwen/DashScope）限制 input 数组最多 64 条；
// OpenAI 等可承受更多，但 64 作为保守默认对所有人都安全，用户可在设置里调大/调小。
const defaultBatchSize = 64

// Embed 批量生成文本向量（OpenAI 兼容 /v1/embeddings）。
// 返回的切片顺序与 inputs 一一对应；inputs 超过批次上限时自动分多次请求，
// 避免触发提供商对单次 input 数组条数的限制（如 "input数组最大不得超过64条"）。
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

	batchSize := setting.BatchSize
	if batchSize <= 0 {
		batchSize = defaultBatchSize
	}

	cfg := openai.DefaultConfig(setting.ApiKey)
	if setting.BaseURL != "" {
		// base_url 按字面量透传，由 go-openai 统一拼接 "/embeddings" 后缀
		// （对齐 OpenAI / go-openai 惯例）。用户应填到 ".../v4"，不要带 /embeddings。
		cfg.BaseURL = setting.BaseURL
	}
	client := openai.NewClientWithConfig(cfg)

	out := make([][]float32, 0, len(inputs))
	for start := 0; start < len(inputs); start += batchSize {
		end := min(start+batchSize, len(inputs))
		batch := inputs[start:end]

		resp, err := client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
			Model: openai.EmbeddingModel(setting.Model),
			Input: batch,
			// 显式请求输出维度，与 vec0 建表维度（setting.Dim）对齐。不传时提供商按模型
			// 原生维度返回（如 Qwen text-embedding-v4 原生 2048），会与按 setting.Dim
			// （如 1024）建的 vec_echo 表冲突，落库报 "Dimension mismatch"。字段带
			// omitempty：Dim 为 0 时自动省略（dimensions 仅 text-embedding-3+ / 兼容模型支持）。
			Dimensions: setting.Dim,
		})
		if err != nil {
			return nil, err
		}
		if len(resp.Data) != len(batch) {
			return nil, ErrEmptyResponse
		}
		for i := range resp.Data {
			out = append(out, resp.Data[i].Embedding)
		}
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
