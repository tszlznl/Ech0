package service

import (
	"context"

	echoService "github.com/lin-snow/ech0/internal/service/echo"
	settingService "github.com/lin-snow/ech0/internal/service/setting"
)

type Service interface {
	GetRecent(ctx context.Context) (string, error)
}

type (
	SettingService = settingService.Service
	EchoService    = echoService.Service
)

type KeyValueRepository interface {
	GetKeyValue(ctx context.Context, key string) (string, error)
	AddOrUpdateKeyValue(ctx context.Context, key, value string) error
	DeleteKeyValue(ctx context.Context, key string) error
}
