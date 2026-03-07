package app

import (
	"context"
	"errors"
	"testing"
)

type mockComponent struct {
	name     string
	order    *[]string
	startErr error
	stopErr  error
}

func (m *mockComponent) Name() string { return m.name }

func (m *mockComponent) Healthy(context.Context) error {
	return nil
}

func (m *mockComponent) Start(context.Context) error {
	*m.order = append(*m.order, "start:"+m.name)
	return m.startErr
}

func (m *mockComponent) Stop(context.Context) error {
	*m.order = append(*m.order, "stop:"+m.name)
	return m.stopErr
}

func TestKernelStartAndStopOrder(t *testing.T) {
	order := make([]string, 0, 8)
	k := NewKernel([]Component{
		&mockComponent{name: "event", order: &order},
		&mockComponent{name: "task", order: &order},
		&mockComponent{name: "http", order: &order},
	})

	if err := k.StartWeb(context.Background()); err != nil {
		t.Fatalf("start web failed: %v", err)
	}
	if err := k.StopWeb(context.Background()); err != nil {
		t.Fatalf("stop web failed: %v", err)
	}

	want := []string{
		"start:event",
		"start:task",
		"start:http",
		"stop:http",
		"stop:task",
		"stop:event",
	}
	if len(order) != len(want) {
		t.Fatalf("unexpected order length: got=%d want=%d", len(order), len(want))
	}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("unexpected order[%d]: got=%q want=%q", i, order[i], want[i])
		}
	}
}

func TestKernelStartRollbackOnFailure(t *testing.T) {
	order := make([]string, 0, 8)
	k := NewKernel([]Component{
		&mockComponent{name: "event", order: &order},
		&mockComponent{name: "task", order: &order, startErr: errors.New("boom")},
		&mockComponent{name: "http", order: &order},
	})

	if err := k.StartWeb(context.Background()); err == nil {
		t.Fatalf("expected start error")
	}
	if k.IsWebRunning() {
		t.Fatalf("web should not be running after rollback")
	}

	want := []string{
		"start:event",
		"start:task",
		"stop:event",
	}
	if len(order) != len(want) {
		t.Fatalf("unexpected order length: got=%d want=%d", len(order), len(want))
	}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("unexpected order[%d]: got=%q want=%q", i, order[i], want[i])
		}
	}
}

func TestKernelStopAll(t *testing.T) {
	order := make([]string, 0, 12)
	k := NewKernel([]Component{
		&mockComponent{name: "event", order: &order},
		&mockComponent{name: "task", order: &order},
	})

	if err := k.StartWeb(context.Background()); err != nil {
		t.Fatalf("start web failed: %v", err)
	}
	if err := k.StopAll(context.Background()); err != nil {
		t.Fatalf("stop all failed: %v", err)
	}

	contains := func(target string) bool {
		for _, item := range order {
			if item == target {
				return true
			}
		}
		return false
	}
	if !contains("stop:task") || !contains("stop:event") {
		t.Fatalf("expected stop order entries missing: %v", order)
	}
}
