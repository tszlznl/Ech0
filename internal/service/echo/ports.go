package service

import (
	"context"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/echo"
	commonService "github.com/lin-snow/ech0/internal/service/common"
)

type Service interface {
	PostEcho(userid uint, newEcho *model.Echo) error
	GetEchosByPage(userid uint, pageQueryDto commonModel.PageQueryDto) (commonModel.PageQueryResult[[]model.Echo], error)
	DeleteEchoById(userid, id uint) error
	GetTodayEchos(userid uint, timezone string) ([]model.Echo, error)
	UpdateEcho(userid uint, echo *model.Echo) error
	LikeEcho(id uint) error
	GetEchoById(userId, id uint) (*model.Echo, error)
	GetAllTags() ([]model.Tag, error)
	DeleteTag(userid, id uint) error
	GetEchosByTagId(userId, tagId uint, pageQueryDto commonModel.PageQueryDto) (commonModel.PageQueryResult[[]model.Echo], error)
}

type CommonService = commonService.Service

type Repository interface {
	CreateEcho(ctx context.Context, newEcho *model.Echo) error
	GetEchosByPage(page, pageSize int, search string, showPrivate bool) ([]model.Echo, int64)
	GetTodayEchos(showPrivate bool, timezone string) []model.Echo
	GetEchosById(ctx context.Context, id uint) (*model.Echo, error)
	UpdateEcho(ctx context.Context, echo *model.Echo) error
	DeleteEchoById(ctx context.Context, id uint) error
	LikeEcho(ctx context.Context, id uint) error
	CreateTag(ctx context.Context, tag *model.Tag) error
	GetAllTags() ([]model.Tag, error)
	GetTagsByNames(ctx context.Context, names []string) ([]*model.Tag, error)
	IncrementTagUsageCount(ctx context.Context, tagID uint) error
	DeleteTagById(ctx context.Context, id uint) error
	GetEchosByTagId(tagID uint, page, pageSize int, search string, showPrivate bool) ([]model.Echo, int64, error)
}
