package event

import (
	"context"
	"testing"

	busen "github.com/lin-snow/Busen"
)

func TestEventBus_StopDrain(t *testing.T) {
	b := NewBus()
	component := NewEventBus(func() *busen.Bus { return b })

	if err := component.Start(context.Background()); err != nil {
		t.Fatalf("start failed: %v", err)
	}
	if err := component.Stop(context.Background()); err != nil {
		t.Fatalf("stop failed: %v", err)
	}
}
