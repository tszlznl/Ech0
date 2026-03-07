package repository

import (
	"context"
	"time"

	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	userModel "github.com/lin-snow/ech0/internal/model/user"
)

type CommonRepositoryInterface interface {
	GetUserByUserId(ctx context.Context, userid uint) (userModel.User, error)
	GetSysAdmin() (userModel.User, error)
	GetAllUsers() ([]userModel.User, error)
	GetAllEchos(showPrivate bool) ([]echoModel.Echo, error)
	GetHeatMap(startUTC, endUTC time.Time) ([]time.Time, error)
}
