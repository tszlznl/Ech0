package service

import (
	"context"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/echo"
	commonService "github.com/lin-snow/ech0/internal/service/common"
	fileService "github.com/lin-snow/ech0/internal/service/file"
)

type Service interface {
	PostEcho(ctx context.Context, newEcho *model.Echo) error
	GetEchosByPage(ctx context.Context, pageQueryDto commonModel.PageQueryDto) (commonModel.PageQueryResult[[]model.Echo], error)
	DeleteEchoById(ctx context.Context, id string) error
	GetTodayEchos(ctx context.Context, timezone string) ([]model.Echo, error)
	UpdateEcho(ctx context.Context, echo *model.Echo) error
	LikeEcho(ctx context.Context, id string) error
	GetEchoById(ctx context.Context, id string) (*model.Echo, error)
	GetAllTags() ([]model.Tag, error)
	DeleteTag(ctx context.Context, id string) error
	GetEchosByTagId(ctx context.Context, tagId string, pageQueryDto commonModel.PageQueryDto) (commonModel.PageQueryResult[[]model.Echo], error)
}

type CommonService = commonService.Service
type FileService = fileService.Service

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
