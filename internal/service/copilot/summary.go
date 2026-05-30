// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/lin-snow/ech0/internal/agent"
	"github.com/lin-snow/ech0/internal/i18n"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"github.com/lin-snow/ech0/pkg/viewer"
	"go.uber.org/zap"
)

// GetRecent 返回站点作者近期活动的 AI 自然语言总结（带缓存 + singleflight 防击穿）。
func (s *CopilotService) GetRecent(ctx context.Context) (string, error) {
	const cacheKey = string(agent.GEN_RECENT)

	if value, ok := s.getRecentFromCache(cacheKey); ok {
		return value, nil
	}

	value, err, _ := s.recentGenGroup.Do(cacheKey, func() (any, error) {
		if cached, ok := s.getRecentFromCache(cacheKey); ok {
			return cached, nil
		}

		output, err := s.buildRecentSummary(ctx)
		if err != nil {
			return "", err
		}

		if err := s.kvRepository.AddOrUpdateKeyValue(ctx, cacheKey, output); err != nil {
			logUtil.GetLogger().
				Error("Failed to add or update key value", zap.Error(err))
		}

		return output, nil
	})
	if err != nil {
		return "", err
	}

	recent, ok := value.(string)
	if !ok {
		return "", errors.New("recent summary type assertion failed")
	}

	return recent, nil
}

func (s *CopilotService) getRecentFromCache(cacheKey string) (string, bool) {
	cachedValue, err := s.kvRepository.GetKeyValue(context.Background(), cacheKey)
	if err != nil {
		return "", false
	}
	return cachedValue, true
}

func (s *CopilotService) buildRecentSummary(ctx context.Context) (string, error) {
	systemCtx := viewer.WithContext(ctx, viewer.NewSystemViewer())
	echos, err := s.echoService.GetEchosByPage(
		systemCtx,
		commonModel.PageQueryDto{
			Page:     1,
			PageSize: 10,
		},
	)
	if err != nil {
		return "", err
	}

	var memos []agent.Message
	for i, e := range echos.Items {
		content := fmt.Sprintf(
			"用户 %s 在 %s 发布了内容 %d ：%s 。 内容标签为：%v。",
			e.Username,
			time.Unix(e.CreatedAt, 0).UTC().Format("2006-01-02 15:04"),
			i+1,
			e.Content,
			e.Tags,
		)

		memos = append(memos, agent.Message{
			Role:    agent.RoleUser,
			Content: content,
		})
	}

	locale := i18n.SystemDefaultLocale()
	in := []agent.Message{
		{
			Role:    agent.RoleSystem,
			Content: summarySystemPromptFor(locale),
		},
		{
			Role:    agent.RoleUser,
			Content: summaryUserPromptFor(locale),
		},
	}

	in = append(in, memos...)

	var setting model.AgentSetting
	if err := s.settingService.GetAgentInfo(&setting); err != nil {
		return "", errors.New(commonModel.AGENT_SETTING_NOT_FOUND)
	}

	output, err := agent.Generate(ctx, setting, in, true)
	if err != nil {
		return "", err
	}

	return output, nil
}
