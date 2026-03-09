package registry

import (
	busen "github.com/lin-snow/Busen"
	"github.com/lin-snow/ech0/internal/config"
	eventbus "github.com/lin-snow/ech0/internal/event/bus"
)

func DeadLetterSubscribeOptions() []busen.SubscribeOption {
	ec := config.Config().Event
	return []busen.SubscribeOption{
		busen.Async(),
		busen.Sequential(),
		busen.WithBuffer(ec.DeadLetterBuffer),
		busen.WithOverflow(eventbus.MapOverflow(ec.DefaultOverflow)),
	}
}

func SystemSubscribeOptions() []busen.SubscribeOption {
	ec := config.Config().Event
	return []busen.SubscribeOption{
		busen.Async(),
		busen.Sequential(),
		busen.WithBuffer(ec.SystemBuffer),
		busen.WithOverflow(eventbus.MapOverflow(ec.DefaultOverflow)),
	}
}

func AgentSubscribeOptions() []busen.SubscribeOption {
	ec := config.Config().Event
	return []busen.SubscribeOption{
		busen.Async(),
		busen.WithParallelism(ec.AgentParallelism),
		busen.WithBuffer(ec.AgentBuffer),
		busen.WithOverflow(eventbus.MapOverflow(ec.DefaultOverflow)),
	}
}

func InboxSubscribeOptions() []busen.SubscribeOption {
	ec := config.Config().Event
	return []busen.SubscribeOption{
		busen.Async(),
		busen.Sequential(),
		busen.WithBuffer(ec.InboxBuffer),
		busen.WithOverflow(eventbus.MapOverflow(ec.DefaultOverflow)),
	}
}
