package event

import (
	busen "github.com/lin-snow/Busen"
	"github.com/lin-snow/ech0/internal/config"
)

func mapOverflow(policy string) busen.OverflowPolicy {
	switch policy {
	case "fail_fast":
		return busen.OverflowFailFast
	case "drop_newest":
		return busen.OverflowDropNewest
	case "drop_oldest":
		return busen.OverflowDropOldest
	default:
		return busen.OverflowBlock
	}
}

func (er *EventRegistrar) deadLetterOptions() []busen.SubscribeOption {
	ec := config.Config().Event
	return []busen.SubscribeOption{
		busen.Async(),
		busen.Sequential(),
		busen.WithBuffer(ec.DeadLetterBuffer),
		busen.WithOverflow(mapOverflow(ec.DefaultOverflow)),
	}
}

func (er *EventRegistrar) systemOptions() []busen.SubscribeOption {
	ec := config.Config().Event
	return []busen.SubscribeOption{
		busen.Async(),
		busen.Sequential(),
		busen.WithBuffer(ec.SystemBuffer),
		busen.WithOverflow(mapOverflow(ec.DefaultOverflow)),
	}
}

func (er *EventRegistrar) agentOptions() []busen.SubscribeOption {
	ec := config.Config().Event
	return []busen.SubscribeOption{
		busen.Async(),
		busen.WithParallelism(ec.AgentParallelism),
		busen.WithBuffer(ec.AgentBuffer),
		busen.WithOverflow(mapOverflow(ec.DefaultOverflow)),
	}
}

func (er *EventRegistrar) inboxOptions() []busen.SubscribeOption {
	ec := config.Config().Event
	return []busen.SubscribeOption{
		busen.Async(),
		busen.Sequential(),
		busen.WithBuffer(ec.InboxBuffer),
		busen.WithOverflow(mapOverflow(ec.DefaultOverflow)),
	}
}
