package event

import (
	"context"

	"github.com/lin-snow/ech0/internal/event"
)

// Runtime 适配 EventRegistrar 到应用生命周期接口。
type Runtime struct {
	registrar *event.EventRegistrar
}

func New(registrar *event.EventRegistrar) *Runtime {
	return &Runtime{registrar: registrar}
}

func (r *Runtime) Name() string {
	return "event"
}

func (r *Runtime) Start(context.Context) error {
	return r.registrar.Register()
}

func (r *Runtime) Stop(context.Context) error {
	r.registrar.Wait()
	return nil
}

func (r *Runtime) Healthy(context.Context) error {
	return nil
}
