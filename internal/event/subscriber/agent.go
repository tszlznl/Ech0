package subscriber

import (
	"context"
	"encoding/json"

	"github.com/lin-snow/ech0/internal/agent"
	contracts "github.com/lin-snow/ech0/internal/event/contracts"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	echoRepository "github.com/lin-snow/ech0/internal/repository/echo"
	inboxRepository "github.com/lin-snow/ech0/internal/repository/inbox"
	keyvalue "github.com/lin-snow/ech0/internal/repository/keyvalue"
	todoRepository "github.com/lin-snow/ech0/internal/repository/todo"
	userRepository "github.com/lin-snow/ech0/internal/repository/user"
)

type AgentProcessor struct {
	echoRepo     echoRepository.EchoRepositoryInterface
	todoRepo     todoRepository.TodoRepositoryInterface
	userRepo     userRepository.UserRepositoryInterface
	keyvalueRepo keyvalue.KeyValueRepositoryInterface
	inboxRepo    inboxRepository.InboxRepositoryInterface
}

func NewAgentProcessor(
	echoRepo echoRepository.EchoRepositoryInterface,
	todoRepo todoRepository.TodoRepositoryInterface,
	userRepo userRepository.UserRepositoryInterface,
	keyvalueRepo keyvalue.KeyValueRepositoryInterface,
	inboxRepo inboxRepository.InboxRepositoryInterface,
) *AgentProcessor {
	return &AgentProcessor{
		echoRepo:     echoRepo,
		todoRepo:     todoRepo,
		userRepo:     userRepo,
		keyvalueRepo: keyvalueRepo,
		inboxRepo:    inboxRepo,
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
