package service

import (
	"context"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	inboxModel "github.com/lin-snow/ech0/internal/model/inbox"
	commonService "github.com/lin-snow/ech0/internal/service/common"
)

type Service interface {
	GetInboxList(userid uint, pageQueryDto commonModel.PageQueryDto) (commonModel.PageQueryResult[[]*inboxModel.Inbox], error)
	GetUnreadInbox(userid uint) ([]*inboxModel.Inbox, error)
	MarkAsRead(userid, inboxID uint) error
	DeleteInbox(userid, inboxID uint) error
	ClearInbox(userid uint) error
}

type CommonService = commonService.Service

type Repository interface {
	GetInboxList(ctx context.Context, offset, limit int, search string) ([]*inboxModel.Inbox, int64, error)
	GetUnreadInbox(ctx context.Context) ([]*inboxModel.Inbox, error)
	GetInboxById(ctx context.Context, id uint) (*inboxModel.Inbox, error)
	UpdateInbox(ctx context.Context, inbox *inboxModel.Inbox) error
	DeleteInbox(ctx context.Context, id uint) error
	ClearInbox(ctx context.Context) error
}
