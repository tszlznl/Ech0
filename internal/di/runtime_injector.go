//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	runtimeHTTP "github.com/lin-snow/ech0/internal/runtime/http"
)

// BuildWebRuntime 构建 HTTP runtime（用于测试和独立启动场景）。
func BuildWebRuntime() (*runtimeHTTP.Runtime, error) {
	wire.Build(
		InfraSet,
		DomainSet,
		ProvideHTTPServer,
		ProvideHTTPRuntime,
	)
	return &runtimeHTTP.Runtime{}, nil
}
