package service

import (
	"context"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/inbox"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	"github.com/lin-snow/ech0/pkg/viewer"
)

type testTransactor struct{}

func (testTransactor) Run(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

type testCommonService struct{}

func (testCommonService) CommonGetUserByUserId(_ context.Context, _ string) (userModel.User, error) {
	return userModel.User{
		ID:      "admin-id",
		IsAdmin: true,
	}, nil
}

func (testCommonService) GetOwner() (userModel.User, error) { return userModel.User{}, nil }
func (testCommonService) GetHeatMap(_ string) ([]commonModel.Heatmap, error) {
	return nil, nil
}
func (testCommonService) GenerateRSS(_ *gin.Context) (string, error) { return "", nil }
func (testCommonService) GetWebsiteTitle(_ string) (string, error)   { return "", nil }

type testInboxRepo struct {
	inbox *model.Inbox
}

func (r *testInboxRepo) GetInboxList(_ context.Context, _, _ int, _ string) ([]*model.Inbox, int64, error) {
	return nil, 0, nil
}

func (r *testInboxRepo) GetUnreadInbox(_ context.Context) ([]*model.Inbox, error) {
	return nil, nil
}

func (r *testInboxRepo) GetInboxById(_ context.Context, _ string) (*model.Inbox, error) {
	return r.inbox, nil
}

func (r *testInboxRepo) UpdateInbox(_ context.Context, inbox *model.Inbox) error {
	r.inbox = inbox
	return nil
}

func (r *testInboxRepo) DeleteInbox(_ context.Context, _ string) error { return nil }
func (r *testInboxRepo) ClearInbox(_ context.Context) error            { return nil }

func TestInboxServiceMarkAsReadSetsReadTrue(t *testing.T) {
	repo := &testInboxRepo{
		inbox: &model.Inbox{
			ID:        "inbox-1",
			Read:      false,
			ReadCount: 0,
			ReadAt:    0,
		},
	}

	svc := NewInboxService(testTransactor{}, testCommonService{}, repo)
	ctx := viewer.WithContext(context.Background(), viewer.NewUserViewer("admin-id"))

	if err := svc.MarkAsRead(ctx, "inbox-1"); err != nil {
		t.Fatalf("mark as read failed: %v", err)
	}

	if !repo.inbox.Read {
		t.Fatalf("expected inbox to be marked read")
	}
	if repo.inbox.ReadCount != 1 {
		t.Fatalf("expected read_count=1, got %d", repo.inbox.ReadCount)
	}
	if repo.inbox.ReadAt <= 0 {
		t.Fatalf("expected read_at > 0")
	}
	if repo.inbox.ReadAt > time.Now().UTC().Unix() {
		t.Fatalf("expected read_at not in the future, got %d", repo.inbox.ReadAt)
	}
}
