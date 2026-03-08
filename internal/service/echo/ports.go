package service

import (
	"context"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/echo"
	commonService "github.com/lin-snow/ech0/internal/service/common"
)

type Service interface {
	PostEcho(userid string, newEcho *model.Echo) error
	GetEchosByPage(userid string, pageQueryDto commonModel.PageQueryDto) (commonModel.PageQueryResult[[]model.Echo], error)
	DeleteEchoById(userid, id string) error
	GetTodayEchos(userid string, timezone string) ([]model.Echo, error)
	UpdateEcho(userid string, echo *model.Echo) error
	LikeEcho(id string) error
	GetEchoById(userId, id string) (*model.Echo, error)
	GetAllTags() ([]model.Tag, error)
	DeleteTag(userid, id string) error
	GetEchosByTagId(userId, tagId string, pageQueryDto commonModel.PageQueryDto) (commonModel.PageQueryResult[[]model.Echo], error)
}

type CommonService = commonService.Service

type Repository interface {
	CreateEcho(ctx context.Context, newEcho *model.Echo) error
	GetEchosByPage(page, pageSize int, search string, showPrivate bool) ([]model.Echo, int64)
	GetTodayEchos(showPrivate bool, timezone string) []model.Echo
	GetEchosById(ctx context.Context, id string) (*model.Echo, error)
	UpdateEcho(ctx context.Context, echo *model.Echo) error
	DeleteEchoById(ctx context.Context, id string) error
	LikeEcho(ctx context.Context, id string) error
	CreateTag(ctx context.Context, tag *model.Tag) error
	GetAllTags() ([]model.Tag, error)
	GetTagsByNames(ctx context.Context, names []string) ([]*model.Tag, error)
	IncrementTagUsageCount(ctx context.Context, tagID string) error
	DeleteTagById(ctx context.Context, id string) error
	GetEchosByTagId(tagID string, page, pageSize int, search string, showPrivate bool) ([]model.Echo, int64, error)
}
