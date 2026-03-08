package bus

import (
	"context"
	"errors"

	busen "github.com/lin-snow/Busen"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"go.uber.org/zap"
)

type Component struct {
	bus *busen.Bus
}

func NewComponent(busProvider func() *busen.Bus) *Component {
	return &Component{bus: busProvider()}
}

func (c *Component) Name() string {
	return "event_bus"
}

func (c *Component) Start(context.Context) error {
	return nil
}

func (c *Component) Stop(ctx context.Context) error {
	if c.bus == nil {
		return errors.New("event bus is nil")
	}
	result, err := c.bus.Shutdown(ctx, busen.ShutdownDrain)
	if err != nil {
		return err
	}
	logUtil.GetLogger().Info("event bus shutdown",
		zap.Bool("completed", result.Completed),
		zap.Int64("processed", result.Processed),
		zap.Int64("dropped", result.Dropped),
		zap.Int64("rejected", result.Rejected))
	return nil
}
