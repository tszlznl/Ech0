package http

import (
	"context"

	"github.com/lin-snow/ech0/internal/server"
)

// Runtime 适配 HTTP Server 到应用生命周期接口。
type Runtime struct {
	server *server.Server
}

func New(server *server.Server) *Runtime {
	return &Runtime{server: server}
}

func (r *Runtime) Name() string {
	return "http"
}

func (r *Runtime) Start(ctx context.Context) error {
	return r.server.Start(ctx)
}

func (r *Runtime) Stop(ctx context.Context) error {
	return r.server.Stop(ctx)
}
