// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"github.com/lin-snow/ech0/internal/storage"
	"golang.org/x/sync/singleflight"
)

// CopilotService 是 Copilot 域的统一服务，同时实现 SummaryService 与 ChatService。
// 近期总结逻辑见 summary.go，Chat 流式问答见 chat.go。
type CopilotService struct {
	echoService    EchoService
	embedding      EmbeddingService
	userReader     UserReader // 取当前对话用户：展示名 + 检索按作者收口
	kvRepository   KeyValueRepository
	storage        *storage.Manager // 多模态：读取命中 Echo 配图字节用于注入模型
	recentGenGroup singleflight.Group
}

var (
	_ SummaryService = (*CopilotService)(nil)
	_ ChatService    = (*CopilotService)(nil)
)

// chatUser 是本轮 Chat 的当前对话用户，用于把检索收口到本人：ID 走 SQL（echos.user_id），
// Username 走向量（embedding 元数据的 username 快照）。展示名同样取自 Username。
type chatUser struct {
	ID       string
	Username string
}

func NewCopilotService(
	echoService EchoService,
	embedding EmbeddingService,
	userReader UserReader,
	kvRepository KeyValueRepository,
	storageManager *storage.Manager,
) *CopilotService {
	return &CopilotService{
		echoService:  echoService,
		embedding:    embedding,
		userReader:   userReader,
		kvRepository: kvRepository,
		storage:      storageManager,
	}
}
