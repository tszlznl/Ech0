package ssh

import (
	"context"
	"time"

	sshRuntime "github.com/lin-snow/ech0/internal/ssh"
)

// Runtime 适配 SSH Server 到应用生命周期接口。
type Runtime struct {
	server *sshRuntime.Server
}

func New(server *sshRuntime.Server) *Runtime {
	return &Runtime{server: server}
}

func (r *Runtime) Name() string {
	return "ssh"
}

func (r *Runtime) Start(ctx context.Context) error {
	return r.server.Start(ctx)
}

func (r *Runtime) Stop(ctx context.Context) error {
	stopCtx := ctx
	var cancel context.CancelFunc
	if stopCtx == nil {
		stopCtx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}
	return r.server.Stop(stopCtx)
}

func (r *Runtime) Healthy(context.Context) error {
	return nil
}
