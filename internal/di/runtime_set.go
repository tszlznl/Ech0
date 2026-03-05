//go:build wireinject
// +build wireinject

package di

import "github.com/google/wire"

var RuntimeSet = wire.NewSet(
	ProvideHTTPServer,
	ProvideHTTPRuntime,
	ProvideEventRuntime,
	ProvideTaskRuntime,
	ProvideCacheRuntime,
)
