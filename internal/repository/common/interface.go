package repository

import (
	"time"

	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	userModel "github.com/lin-snow/ech0/internal/model/user"
)

type CommonRepositoryInterface interface {
	GetUserByUserId(userid uint) (userModel.User, error)
	GetSysAdmin() (userModel.User, error)
	GetAllUsers() ([]userModel.User, error)
	GetAllEchos(showPrivate bool) ([]echoModel.Echo, error)
	GetHeatMap(startUTC, endUTC time.Time) ([]time.Time, error)
}
