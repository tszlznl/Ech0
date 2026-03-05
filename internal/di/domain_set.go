//go:build wireinject
// +build wireinject

package di

import "github.com/google/wire"

var DomainSet = wire.NewSet(
	ProvideHandlers,
	ProvideTasker,
	ProvideEventRegistrar,
)
