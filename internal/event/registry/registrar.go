package registry

import (
	"context"
	"sync/atomic"

	busen "github.com/lin-snow/Busen"
	contracts "github.com/lin-snow/ech0/internal/event/contracts"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"go.uber.org/zap"
)

type WebhookObserver interface {
	HandleObservation(ctx context.Context, obs contracts.WebhookObservation) error
	Stop()
	Wait()
}

type DeadLetterHandler interface {
	Handle(ctx context.Context, event contracts.DeadLetterRetriedEvent) error
}

type BackupScheduleHandler interface {
	HandleBackupScheduleUpdated(ctx context.Context, e contracts.UpdateBackupScheduleEvent) error
}

type AgentEventHandler interface {
	HandleEchoCreated(ctx context.Context, e contracts.EchoCreatedEvent) error
	HandleEchoUpdated(ctx context.Context, e contracts.EchoUpdatedEvent) error
	HandleUserDeleted(ctx context.Context, e contracts.UserDeletedEvent) error
}

type InboxEventHandler interface {
	HandleEch0UpdateCheck(ctx context.Context, e contracts.Ech0UpdateCheckEvent) error
	HandleInboxClear(ctx context.Context, e contracts.InboxClearEvent) error
}

type EventHandlers struct {
	wbd WebhookObserver
	dlr DeadLetterHandler
	bs  BackupScheduleHandler
	ap  AgentEventHandler
	id  InboxEventHandler
}

func NewEventHandlers(
	wbd WebhookObserver,
	dlr DeadLetterHandler,
	bs BackupScheduleHandler,
	ap AgentEventHandler,
	id InboxEventHandler,
) *EventHandlers {
	return &EventHandlers{wbd: wbd, dlr: dlr, bs: bs, ap: ap, id: id}
}

type EventRegistrar struct {
	bus        *busen.Bus
	eh         *EventHandlers
	unsub      []func()
	registered atomic.Bool
}

func NewEventRegistry(busProvider func() *busen.Bus, eh *EventHandlers) *EventRegistrar {
	return &EventRegistrar{bus: busProvider(), eh: eh}
}

func (er *EventRegistrar) Register() error {
	if er.registered.Load() {
		return nil
	}

	if err := er.registerDeadLetter(); err != nil {
		return err
	}
	if err := er.registerSystem(); err != nil {
		return err
	}
	if err := er.registerAgent(); err != nil {
		return err
	}
	if err := er.registerInbox(); err != nil {
		return err
	}

	err := er.bus.UseObserver(
		func(ctx context.Context, obs busen.Observation) {
			if obs.Topic == contracts.TopicDeadLetterRetried {
				return
			}
			evt, err := contracts.NewWebhookObservation(obs.Topic, obs.Value, obs.Meta)
			if err != nil {
				logUtil.GetLogger().Warn("build webhook observation failed",
					zap.String("topic", obs.Topic),
					zap.Error(err))
				return
			}
			if err := er.eh.wbd.HandleObservation(ctx, evt); err != nil {
				logUtil.GetLogger().Warn("dispatch webhook observation failed",
					zap.String("topic", obs.Topic),
					zap.Error(err))
			}
		},
	)
	if err != nil {
		return err
	}

	er.registered.Store(true)
	return nil
}

func (er *EventRegistrar) Stop() error {
	if !er.registered.Load() {
		return nil
	}
	for i := len(er.unsub) - 1; i >= 0; i-- {
		er.unsub[i]()
	}
	er.unsub = nil
	er.eh.wbd.Stop()
	er.eh.wbd.Wait()
	return nil
}
