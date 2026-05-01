// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package agent

import (
	"context"
	"errors"
	"fmt"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/packages/param"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
	"github.com/sashabaranov/go-openai"
	"google.golang.org/genai"
)

const (
	GEN_RECENT = "gen_recent"

	// anthropicDefaultMaxTokens 是 Anthropic API 必填字段 max_tokens 的默认值
	anthropicDefaultMaxTokens = 4096
)

// Role 表示一条对话消息的发送者角色
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// Message 是 Agent 内部使用的对话消息抽象
type Message struct {
	Role    Role
	Content string
}

// Generate 调用配置的 LLM 提供商生成回复
func Generate(
	ctx context.Context,
	setting model.AgentSetting,
	in []Message,
	usePrompt bool,
	temperature ...float32,
) (string, error) {
	if !setting.Enable {
		return "", errors.New(commonModel.AGENT_NOT_ENABLED)
	}
	if setting.Model == "" {
		return "", errors.New(commonModel.AGENT_MODEL_MISSING)
	}
	if setting.Provider == "" {
		return "", errors.New(commonModel.AGENT_PROVIDER_NOT_FOUND)
	}
	if setting.ApiKey == "" && setting.Provider != string(commonModel.OpenAI) {
		// OpenAI 兼容场景下 Ollama 等本地服务允许空 ApiKey
		return "", errors.New(commonModel.AGENT_API_KEY_MISSING)
	}

	if setting.Prompt != "" && usePrompt {
		// 维持历史行为：自定义 Prompt 以 user 消息追加在末尾
		in = append(in, Message{Role: RoleUser, Content: setting.Prompt})
	}

	var t *float32
	if len(temperature) > 0 {
		t = &temperature[0]
	}

	switch setting.Provider {
	case string(commonModel.OpenAI):
		return generateOpenAI(ctx, setting, in, t)
	case string(commonModel.Anthropic):
		return generateAnthropic(ctx, setting, in, t)
	case string(commonModel.Gemini):
		return generateGemini(ctx, setting, in, t)
	default:
		return "", errors.New(commonModel.AGENT_PROVIDER_NOT_FOUND)
	}
}

func generateOpenAI(
	ctx context.Context,
	setting model.AgentSetting,
	in []Message,
	t *float32,
) (string, error) {
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
	}
	if t != nil {
		req.Temperature = *t
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", errors.New("openai: empty response")
	}
	return resp.Choices[0].Message.Content, nil
}

func toOpenAIRole(r Role) string {
	switch r {
	case RoleSystem:
		return openai.ChatMessageRoleSystem
	case RoleAssistant:
		return openai.ChatMessageRoleAssistant
	default:
		return openai.ChatMessageRoleUser
	}
}

func generateAnthropic(
	ctx context.Context,
	setting model.AgentSetting,
	in []Message,
	t *float32,
) (string, error) {
	opts := []option.RequestOption{option.WithAPIKey(setting.ApiKey)}
	if setting.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(setting.BaseURL))
	}
	client := anthropic.NewClient(opts...)

	var systemBlocks []anthropic.TextBlockParam
	msgs := make([]anthropic.MessageParam, 0, len(in))
	for _, m := range in {
		switch m.Role {
		case RoleSystem:
			// Anthropic 把 system 提到顶层 system 字段，不作为消息
			systemBlocks = append(systemBlocks, anthropic.TextBlockParam{Text: m.Content})
		case RoleAssistant:
			msgs = append(msgs, anthropic.NewAssistantMessage(anthropic.NewTextBlock(m.Content)))
		default:
			msgs = append(msgs, anthropic.NewUserMessage(anthropic.NewTextBlock(m.Content)))
		}
	}

	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(setting.Model),
		MaxTokens: anthropicDefaultMaxTokens,
		Messages:  msgs,
	}
	if len(systemBlocks) > 0 {
		params.System = systemBlocks
	}
	if t != nil {
		params.Temperature = param.NewOpt(float64(*t))
	}

	resp, err := client.Messages.New(ctx, params)
	if err != nil {
		return "", err
	}

	var out string
	for _, block := range resp.Content {
		if block.Type == "text" {
			out += block.Text
		}
	}
	if out == "" {
		return "", errors.New("anthropic: empty text response")
	}
	return out, nil
}

func generateGemini(
	ctx context.Context,
	setting model.AgentSetting,
	in []Message,
	t *float32,
) (string, error) {
	cfg := &genai.ClientConfig{
		APIKey:  setting.ApiKey,
		Backend: genai.BackendGeminiAPI,
	}
	if setting.BaseURL != "" {
		cfg.HTTPOptions = genai.HTTPOptions{BaseURL: setting.BaseURL}
	}
	client, err := genai.NewClient(ctx, cfg)
	if err != nil {
		return "", err
	}

	var systemContent *genai.Content
	contents := make([]*genai.Content, 0, len(in))
	for _, m := range in {
		switch m.Role {
		case RoleSystem:
			// Gemini 通过 GenerateContentConfig.SystemInstruction 传递；多条 system 拼接
			if systemContent == nil {
				systemContent = &genai.Content{Parts: []*genai.Part{{Text: m.Content}}}
			} else {
				systemContent.Parts = append(systemContent.Parts, &genai.Part{Text: m.Content})
			}
		case RoleAssistant:
			contents = append(contents, &genai.Content{
				Role:  genai.RoleModel,
				Parts: []*genai.Part{{Text: m.Content}},
			})
		default:
			contents = append(contents, &genai.Content{
				Role:  genai.RoleUser,
				Parts: []*genai.Part{{Text: m.Content}},
			})
		}
	}

	genCfg := &genai.GenerateContentConfig{}
	if systemContent != nil {
		genCfg.SystemInstruction = systemContent
	}
	if t != nil {
		genCfg.Temperature = t
	}

	resp, err := client.Models.GenerateContent(ctx, setting.Model, contents, genCfg)
	if err != nil {
		return "", err
	}
	out := resp.Text()
	if out == "" {
		return "", fmt.Errorf("gemini: empty text response")
	}
	return out, nil
}
