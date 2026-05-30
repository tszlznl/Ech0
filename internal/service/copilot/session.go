// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"encoding/json"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	embeddingModel "github.com/lin-snow/ech0/internal/model/embedding"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"github.com/lin-snow/ech0/pkg/viewer"
	"go.uber.org/zap"
)

// maxStoredChatMessages 是单个用户持久化会话保留的最大消息条数（超出取最近 N 条）。
const maxStoredChatMessages = 50

// ChatMessage 是持久化在 KeyValue 里的一条聊天消息（仅做展示恢复，模型不读历史）。
type ChatMessage struct {
	Role    string                        `json:"role"`
	Content string                        `json:"content"`
	Sources []embeddingModel.SearchResult `json:"sources,omitempty"`
}

// chatSessionKey 按 userID 生成会话键（每个用户一条会话）。
func chatSessionKey(userID string) string {
	return commonModel.ChatSessionKeyPrefix + userID
}

// loadSession 读取某用户的持久化会话；userID 为空、未命中或解析失败都返回 nil（best-effort）。
func (s *CopilotService) loadSession(ctx context.Context, userID string) []ChatMessage {
	if userID == "" {
		return nil
	}
	raw, err := s.kvRepository.GetKeyValue(ctx, chatSessionKey(userID))
	if err != nil {
		return nil
	}
	var msgs []ChatMessage
	if err := json.Unmarshal([]byte(raw), &msgs); err != nil {
		return nil
	}
	return msgs
}

// appendTurn 把本轮消息追加进用户会话并落盘（封顶最近 maxStoredChatMessages 条）。
// userID 为空直接跳过；任何失败仅告警，不返回错误（best-effort，不影响主流程）。
func (s *CopilotService) appendTurn(ctx context.Context, userID string, turn ...ChatMessage) {
	if userID == "" {
		return
	}
	msgs := append(s.loadSession(ctx, userID), turn...)
	if len(msgs) > maxStoredChatMessages {
		msgs = msgs[len(msgs)-maxStoredChatMessages:]
	}
	payload, err := json.Marshal(msgs)
	if err != nil {
		logUtil.GetLogger().Warn("failed to marshal chat session",
			zap.String("module", "copilot"), zap.Error(err))
		return
	}
	if err := s.kvRepository.AddOrUpdateKeyValue(ctx, chatSessionKey(userID), string(payload)); err != nil {
		logUtil.GetLogger().Warn("failed to persist chat session",
			zap.String("module", "copilot"), zap.Error(err))
	}
}

// persistTurn 在一轮问答正常收尾时把 user/assistant 两条消息追加进持久化会话。
// assistant 内容为空（如模型未产出文本）也照常落盘，与前端展示保持一致。
func (s *CopilotService) persistTurn(ctx context.Context, userID, question, answer string, sources []embeddingModel.SearchResult) {
	s.appendTurn(ctx, userID,
		ChatMessage{Role: "user", Content: question},
		ChatMessage{Role: "assistant", Content: answer, Sources: sources},
	)
}

// GetSession 返回当前登录用户的持久化会话（无会话时返回空切片，便于前端拿到数组）。
func (s *CopilotService) GetSession(ctx context.Context) ([]ChatMessage, error) {
	userID := viewer.MustFromContext(ctx).UserID()
	msgs := s.loadSession(ctx, userID)
	if msgs == nil {
		return []ChatMessage{}, nil
	}
	return msgs, nil
}

// ClearSession 删除当前登录用户的持久化会话（不可恢复，由前端二次确认）。
func (s *CopilotService) ClearSession(ctx context.Context) error {
	userID := viewer.MustFromContext(ctx).UserID()
	if userID == "" {
		return nil
	}
	return s.kvRepository.DeleteKeyValue(ctx, chatSessionKey(userID))
}
