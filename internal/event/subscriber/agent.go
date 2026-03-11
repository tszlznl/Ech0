package subscriber

import (
	"context"
	"encoding/json"

	"github.com/lin-snow/ech0/internal/agent"
	contracts "github.com/lin-snow/ech0/internal/event/contracts"
	registry "github.com/lin-snow/ech0/internal/event/registry"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	agentService "github.com/lin-snow/ech0/internal/service/agent"
)

type AgentProcessor struct {
	keyvalueRepo agentService.KeyValueRepository
}

func NewAgentProcessor(
	keyvalueRepo agentService.KeyValueRepository,
) *AgentProcessor {
	return &AgentProcessor{
		keyvalueRepo: keyvalueRepo,
	}
}

func (ap *AgentProcessor) handle(ctx context.Context) error {
	var agentSetting settingModel.AgentSetting
	if agentSettingStr, err := ap.keyvalueRepo.GetKeyValue(ctx, commonModel.AgentSettingKey); err == nil {
		if err := json.Unmarshal([]byte(agentSettingStr), &agentSetting); err != nil {
			return err
		}
	}

	return ap.clearCache()
}

func (ap *AgentProcessor) HandleEchoCreated(ctx context.Context, e contracts.EchoCreatedEvent) error {
	_ = e
	return ap.handle(ctx)
}

func (ap *AgentProcessor) HandleEchoUpdated(ctx context.Context, e contracts.EchoUpdatedEvent) error {
	_ = e
	return ap.handle(ctx)
}

func (ap *AgentProcessor) HandleUserDeleted(ctx context.Context, e contracts.UserDeletedEvent) error {
	_ = e
	return ap.handle(ctx)
}

func (ap *AgentProcessor) clearCache() error {
	return ap.keyvalueRepo.DeleteKeyValue(context.Background(), string(agent.GEN_RECENT))
}

func (ap *AgentProcessor) Subscriptions() []registry.Subscription {
	return []registry.Subscription{
		registry.TopicSubscription(
			contracts.TopicEchoCreated,
			ap.HandleEchoCreated,
			registry.AgentSubscribeOptions()...,
		),
		registry.TopicSubscription(
			contracts.TopicEchoUpdated,
			ap.HandleEchoUpdated,
			registry.AgentSubscribeOptions()...,
		),
		registry.TopicSubscription(
			contracts.TopicUserDeleted,
			ap.HandleUserDeleted,
			registry.AgentSubscribeOptions()...,
		),
	}
}
