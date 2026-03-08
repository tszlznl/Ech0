package service

import (
	"context"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	inboxModel "github.com/lin-snow/ech0/internal/model/inbox"
	commonService "github.com/lin-snow/ech0/internal/service/common"
)

type Service interface {
	GetInboxList(userid string, pageQueryDto commonModel.PageQueryDto) (commonModel.PageQueryResult[[]*inboxModel.Inbox], error)
	GetUnreadInbox(userid string) ([]*inboxModel.Inbox, error)
	MarkAsRead(userid, inboxID string) error
	DeleteInbox(userid, inboxID string) error
	ClearInbox(userid string) error
}

type CommonService = commonService.Service

type Repository interface {
	GetInboxList(ctx context.Context, offset, limit int, search string) ([]*inboxModel.Inbox, int64, error)
	GetUnreadInbox(ctx context.Context) ([]*inboxModel.Inbox, error)
	GetInboxById(ctx context.Context, id string) (*inboxModel.Inbox, error)
	UpdateInbox(ctx context.Context, inbox *inboxModel.Inbox) error
	DeleteInbox(ctx context.Context, id string) error
	ClearInbox(ctx context.Context) error
}
