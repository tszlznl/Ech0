// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import "golang.org/x/sync/singleflight"

// CopilotService 是 Copilot 域的统一服务，同时实现 SummaryService 与 ChatService。
// 近期总结逻辑见 summary.go，Chat 流式问答见 chat.go。
type CopilotService struct {
	settingService SettingService
	echoService    EchoService
	embedding      EmbeddingService
	kvRepository   KeyValueRepository
	recentGenGroup singleflight.Group
}

var (
	_ SummaryService = (*CopilotService)(nil)
	_ ChatService    = (*CopilotService)(nil)
)

func NewCopilotService(
	settingService SettingService,
	echoService EchoService,
	embedding EmbeddingService,
	kvRepository KeyValueRepository,
) *CopilotService {
	return &CopilotService{
		settingService: settingService,
		echoService:    echoService,
		embedding:      embedding,
		kvRepository:   kvRepository,
	}
}
