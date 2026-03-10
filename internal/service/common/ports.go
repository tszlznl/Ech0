package service

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	userModel "github.com/lin-snow/ech0/internal/model/user"
)

type Service interface {
	CommonGetUserByUserId(ctx context.Context, userId string) (userModel.User, error)
	GetOwner() (userModel.User, error)
	GetHeatMap(timezone string) ([]commonModel.Heatmap, error)
	GenerateRSS(ctx *gin.Context) (string, error)
	GetWebsiteTitle(websiteURL string) (string, error)
}

type CommonRepository interface {
	GetUserByUserId(ctx context.Context, id string) (userModel.User, error)
	GetOwner(ctx context.Context) (userModel.User, error)
	GetHeatMap(ctx context.Context, startTime, endTime time.Time) ([]time.Time, error)
}
