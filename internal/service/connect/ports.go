package service

import (
	"context"

	model "github.com/lin-snow/ech0/internal/model/connect"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	commonService "github.com/lin-snow/ech0/internal/service/common"
	settingService "github.com/lin-snow/ech0/internal/service/setting"
)

type Service interface {
	AddConnect(userid uint, connected model.Connected) error
	DeleteConnect(userid, id uint) error
	GetConnect() (model.Connect, error)
	GetConnectsInfo() ([]model.Connect, error)
	GetConnects() ([]model.Connected, error)
}

type Repository interface {
	GetAllConnects(ctx context.Context) ([]model.Connected, error)
	CreateConnect(ctx context.Context, connected *model.Connected) error
	DeleteConnect(ctx context.Context, id uint) error
}

type EchoRepository interface {
	GetTodayEchos(showPrivate bool, timezone string) []echoModel.Echo
}

type (
	CommonService  = commonService.Service
	SettingService = settingService.Service
)
