package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/lin-snow/ech0/internal/agent"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"github.com/lin-snow/ech0/pkg/viewer"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
)

type AgentService struct {
	settingService SettingService
	echoService    EchoService
	kvRepository   KeyValueRepository
	recentGenGroup singleflight.Group
}

func NewAgentService(
	settingService SettingService,
	echoService EchoService,
	kvRepository KeyValueRepository,
) *AgentService {
	return &AgentService{
		settingService: settingService,
		echoService:    echoService,
		kvRepository:   kvRepository,
	}
}

func (agentService *AgentService) GetRecent(ctx context.Context) (string, error) {
	const cacheKey = string(agent.GEN_RECENT)

	if value, ok := agentService.getRecentFromCache(cacheKey); ok {
		return value, nil
	}

	value, err, _ := agentService.recentGenGroup.Do(cacheKey, func() (any, error) {
		if cached, ok := agentService.getRecentFromCache(cacheKey); ok {
			return cached, nil
		}

		output, err := agentService.buildRecentSummary(ctx)
		if err != nil {
			return "", err
		}

		if err := agentService.kvRepository.AddOrUpdateKeyValue(ctx, cacheKey, output); err != nil {
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

func (agentService *AgentService) getRecentFromCache(cacheKey string) (string, bool) {
	cachedValue, err := agentService.kvRepository.GetKeyValue(context.Background(), cacheKey)
	if err != nil {
		return "", false
	}
	return cachedValue, true
}

func (agentService *AgentService) buildRecentSummary(ctx context.Context) (string, error) {
	systemCtx := viewer.WithContext(ctx, viewer.NewSystemViewer())
	echos, err := agentService.echoService.GetEchosByPage(
		systemCtx,
		commonModel.PageQueryDto{
			Page:     1,
			PageSize: 10,
		},
	)
	if err != nil {
		return "", err
	}

	var memos []*schema.Message
	for i, e := range echos.Items {
		content := fmt.Sprintf(
			"用户 %s 在 %s 发布了内容 %d ：%s 。 内容标签为：%v。",
			e.Username,
			time.Unix(e.CreatedAt, 0).UTC().Format("2006-01-02 15:04"),
			i+1,
			e.Content,
			e.Tags,
		)

		memos = append(memos, &schema.Message{
			Role:    schema.User,
			Content: content,
		})
	}

	in := []*schema.Message{
		{
			Role: schema.System,
			Content: `
				这是“近况总结”场景，请使用简洁自然的中文表达。
				不使用复杂格式：不要标题、列表、表格、代码块、链接。
				不要输出任何原始 HTML 标签。
				可使用纯文字、Emoji 和正常换行来增强可读性。
				回复保持简洁，聚焦作者最近的活动和状态。`,
		},
		{
			Role:    schema.User,
			Content: "请根据提供的近期互动内容（内容可能包括日常生活、句子诗词摘抄、吐槽等等），总结该用户最近的活动和状态，突出作者状态即可，不需要详细描述内容，如果没有任何内容，请回复作者最近很神秘~",
		},
	}

	in = append(in, memos...)

	var setting model.AgentSetting
	if err := agentService.settingService.GetAgentInfo(&setting); err != nil {
		return "", errors.New(commonModel.AGENT_SETTING_NOT_FOUND)
	}

	output, err := agent.Generate(ctx, setting, in, true)
	if err != nil {
		return "", err
	}

	return output, nil
}
