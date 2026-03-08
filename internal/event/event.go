package event

import (
	"reflect"

	busen "github.com/lin-snow/Busen"
	"github.com/lin-snow/ech0/internal/config"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"go.uber.org/zap"
)

const (
	MetaKeySource    = "source"
	MetaKeyActorID   = "actor_id"
	MetaKeyTraceID   = "trace_id"
	MetaKeyRequestID = "request_id"
)

func NewBus() *busen.Bus {
	ec := config.Config().Event
	hooks := busen.Hooks{
		OnHandlerError: func(info busen.HandlerError) {
			logUtil.GetLogger().Error("busen handler error",
				zap.String("event_type", safeTypeString(info.EventType)),
				zap.String("topic", info.Topic),
				zap.String("key", info.Key),
				zap.Bool("async", info.Async),
				zap.Error(info.Err))
		},
		OnHandlerPanic: func(info busen.HandlerPanic) {
			logUtil.GetLogger().Error("busen handler panic",
				zap.String("event_type", safeTypeString(info.EventType)),
				zap.String("topic", info.Topic),
				zap.String("key", info.Key),
				zap.Bool("async", info.Async),
				zap.Any("panic", info.Value))
		},
		OnPublishDone: func(info busen.PublishDone) {
			if info.Err == nil {
				return
			}
			logUtil.GetLogger().Warn("busen publish done with errors",
				zap.String("event_type", safeTypeString(info.EventType)),
				zap.String("topic", info.Topic),
				zap.String("key", info.Key),
				zap.Int("matched_subscribers", info.MatchedSubscribers),
				zap.Int("delivered_subscribers", info.DeliveredSubscribers),
				zap.Error(info.Err))
		},
		OnEventDropped: func(info busen.DroppedEvent) {
			logUtil.GetLogger().Warn("busen event dropped",
				zap.String("event_type", safeTypeString(info.EventType)),
				zap.String("topic", info.Topic),
				zap.String("key", info.Key),
				zap.Int("queue_len", info.QueueLen),
				zap.Int("queue_cap", info.QueueCap),
				zap.Error(info.Reason))
		},
		OnEventRejected: func(info busen.RejectedEvent) {
			logUtil.GetLogger().Warn("busen event rejected",
				zap.String("event_type", safeTypeString(info.EventType)),
				zap.String("topic", info.Topic),
				zap.String("key", info.Key),
				zap.Int("queue_len", info.QueueLen),
				zap.Int("queue_cap", info.QueueCap),
				zap.Error(info.Reason))
		},
	}

	b := busen.New(
		busen.WithDefaultBuffer(ec.DefaultBuffer),
		busen.WithDefaultOverflow(mapOverflow(ec.DefaultOverflow)),
		busen.WithHooks(hooks),
		busen.WithMetadataBuilder(func(input busen.PublishMetadataInput) map[string]string {
			return map[string]string{
				"source": "ech0",
			}
		}),
	)

	return b
}

func safeTypeString(t reflect.Type) string {
	if t == nil {
		return ""
	}
	return t.String()
}

func eventNameOf(payload any) string {
	if payload == nil {
		return ""
	}
	t := reflect.TypeOf(payload)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Name() != "" {
		return t.Name()
	}
	return t.String()
}
