//go:build wireinject
// +build wireinject

package di

import "github.com/google/wire"

var AppSet = wire.NewSet(
	ProvideWebComponents,
	ProvideApp,
)
