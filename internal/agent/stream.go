// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package agent

import (
	"context"
	"errors"
	"io"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
	openai "github.com/sashabaranov/go-openai"
)

// StreamChunk 是流式生成的一个增量片段；Err 非空表示出错（之后 channel 会关闭）。
type StreamChunk struct {
	Delta string
	Err   error
}

// GenerateStream 以流式方式调用配置的 LLM，返回只读 channel，结束时关闭。
//
// OpenAI 兼容协议为真流式（覆盖 OpenAI / DeepSeek / Qwen / Ollama 等绝大多数场景）；
// Anthropic / Gemini 在 v1 先以"一次性生成后整段作为单个 chunk 返回"的方式回退，
// 保证行为一致与可用，后续可补真流式（见 docs/dev/llm-chat-design.md §6.6）。
func GenerateStream(
	ctx context.Context,
	setting model.AgentSetting,
	in []Message,
	usePrompt bool,
	temperature ...float32,
) (<-chan StreamChunk, error) {
	if !setting.Enable {
		return nil, errors.New(commonModel.AGENT_NOT_ENABLED)
	}
	if setting.Model == "" {
		return nil, errors.New(commonModel.AGENT_MODEL_MISSING)
	}
	if setting.Protocol == "" {
		return nil, errors.New(commonModel.AGENT_PROTOCOL_NOT_FOUND)
	}
	if setting.ApiKey == "" && setting.Protocol != string(commonModel.OpenAI) {
		return nil, errors.New(commonModel.AGENT_API_KEY_MISSING)
	}

	if setting.Prompt != "" && usePrompt {
		in = append(in, Message{Role: RoleUser, Content: setting.Prompt})
	}

	var t *float32
	if len(temperature) > 0 {
		t = &temperature[0]
	}

	ch := make(chan StreamChunk)

	switch setting.Protocol {
	case string(commonModel.OpenAI):
		go streamOpenAI(ctx, setting, in, t, ch)
	default:
		go func() {
			defer close(ch)
			out, err := generateNonStream(ctx, setting, in, t)
			if err != nil {
				ch <- StreamChunk{Err: err}
				return
			}
			select {
			case ch <- StreamChunk{Delta: out}:
			case <-ctx.Done():
			}
		}()
	}

	return ch, nil
}

// generateNonStream 复用现有的非流式实现，用于 Anthropic / Gemini 的 v1 回退。
func generateNonStream(
	ctx context.Context,
	setting model.AgentSetting,
	in []Message,
	t *float32,
) (string, error) {
	switch setting.Protocol {
	case string(commonModel.Anthropic):
		return generateAnthropic(ctx, setting, in, t)
	case string(commonModel.Gemini):
		return generateGemini(ctx, setting, in, t)
	default:
		return "", errors.New(commonModel.AGENT_PROTOCOL_NOT_FOUND)
	}
}

func streamOpenAI(
	ctx context.Context,
	setting model.AgentSetting,
	in []Message,
	t *float32,
	ch chan<- StreamChunk,
) {
	defer close(ch)

	cfg := openai.DefaultConfig(setting.ApiKey)
	if setting.BaseURL != "" {
		cfg.BaseURL = setting.BaseURL
	}
	client := openai.NewClientWithConfig(cfg)

	msgs := make([]openai.ChatCompletionMessage, 0, len(in))
	for _, m := range in {
		msgs = append(msgs, openai.ChatCompletionMessage{
			Role:    toOpenAIRole(m.Role),
			Content: m.Content,
		})
	}

	req := openai.ChatCompletionRequest{
		Model:    setting.Model,
		Messages: msgs,
		Stream:   true,
	}
	if t != nil {
		req.Temperature = *t
	}

	stream, err := client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		ch <- StreamChunk{Err: err}
		return
	}
	defer stream.Close()

	for {
		resp, recvErr := stream.Recv()
		if errors.Is(recvErr, io.EOF) {
			return
		}
		if recvErr != nil {
			ch <- StreamChunk{Err: recvErr}
			return
		}
		if len(resp.Choices) == 0 {
			continue
		}
		delta := resp.Choices[0].Delta.Content
		if delta == "" {
			continue
		}
		select {
		case ch <- StreamChunk{Delta: delta}:
		case <-ctx.Done():
			return
		}
	}
}
