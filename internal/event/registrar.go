package event

import (
	"context"
	"sync/atomic"

	busen "github.com/lin-snow/Busen"
)

// EventHandlers 事件处理器集合
type EventHandlers struct {
	wbd *WebhookDispatcher  // webhook 事件处理器
	dlr *DeadLetterResolver // 死信处理器
	bs  *BackupScheduler    // 备份事件调度器
	ap  *AgentProcessor     // Agent事件处理器
	id  *InboxDispatcher    // Inbox事件处理器
}

// NewEventHandlers 创建一个新的事件处理器集合
func NewEventHandlers(
	wbd *WebhookDispatcher,
	dlr *DeadLetterResolver,
	bs *BackupScheduler,
	ap *AgentProcessor,
	id *InboxDispatcher,
) *EventHandlers {
	return &EventHandlers{wbd: wbd, dlr: dlr, bs: bs, ap: ap, id: id}
}

// EventRegistrar 事件注册器
type EventRegistrar struct {
	bus        *busen.Bus     // 事件总线
	eh         *EventHandlers // 事件处理器集合
	unsub      []func()
	registered atomic.Bool
}

// NewEventRegistry 创建一个新的事件注册表
func NewEventRegistry(busProvider func() *busen.Bus, eh *EventHandlers) *EventRegistrar {
	return &EventRegistrar{bus: busProvider(), eh: eh}
}

// Register 注册事件处理函数
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
			switch obs.Topic {
			case TopicDeadLetterRetried:
				return
			}
			evt, err := newWebhookObservation(obs.Topic, obs.Value, obs.Meta)
			if err != nil {
				return
			}
			_ = er.eh.wbd.HandleObservation(ctx, evt)
		},
	)
	if err != nil {
		return err
	}

	er.registered.Store(true)
	return nil
}

// Stop 等待已投递的异步处理任务完成。
func (er *EventRegistrar) Stop() error {
	if !er.registered.Load() {
		return nil
	}
	for i := len(er.unsub) - 1; i >= 0; i-- {
		er.unsub[i]()
	}
	er.unsub = nil

	er.eh.wbd.Wait()
	return nil
}
