package subscriber

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	contracts "github.com/lin-snow/ech0/internal/event/contracts"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	inboxModel "github.com/lin-snow/ech0/internal/model/inbox"
	inboxRepository "github.com/lin-snow/ech0/internal/repository/inbox"
	keyvalueRepository "github.com/lin-snow/ech0/internal/repository/keyvalue"
	githubUtil "github.com/lin-snow/ech0/internal/util/github"
	"golang.org/x/mod/semver"
	"gorm.io/gorm"
)

type InboxDispatcher struct {
	inboxRepo    inboxRepository.InboxRepositoryInterface
	keyvalueRepo keyvalueRepository.KeyValueRepositoryInterface
}

func NewInboxDispatcher(inboxRepo inboxRepository.InboxRepositoryInterface, keyvalueRepo keyvalueRepository.KeyValueRepositoryInterface) *InboxDispatcher {
	return &InboxDispatcher{inboxRepo: inboxRepo, keyvalueRepo: keyvalueRepo}
}

func (id *InboxDispatcher) HandleEch0UpdateCheck(ctx context.Context, _ contracts.Ech0UpdateCheckEvent) error {
	return id.handleEch0UpdateCheck(ctx)
}

func (id *InboxDispatcher) HandleInboxClear(ctx context.Context, _ contracts.InboxClearEvent) error {
	return id.handleInboxClear(ctx)
}

func (id *InboxDispatcher) handleEch0UpdateCheck(ctx context.Context) error {
	currentVersion := commonModel.Version
	latestVersion, err := githubUtil.GetLatestVersion()
	if err != nil {
		return err
	}

	releaseVersion, err := id.keyvalueRepo.GetKeyValue(ctx, commonModel.ReleaseVersionKey)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if releaseVersion != "" && releaseVersion == latestVersion {
		return nil
	}

	cur := semver.Canonical("v" + strings.TrimPrefix(strings.TrimSpace(currentVersion), "v"))
	lat := semver.Canonical("v" + strings.TrimPrefix(strings.TrimSpace(latestVersion), "v"))
	if cur != "" && lat != "" && semver.Compare(lat, cur) > 0 {
		meta, _ := json.Marshal(map[string]string{
			"latest_version":  latestVersion,
			"current_version": currentVersion,
		})
		err = id.inboxRepo.PostInbox(ctx, &inboxModel.Inbox{
			Source:    string(commonModel.SystemSource),
			Content:   fmt.Sprintf("有新版本可用，请更新：%s", latestVersion),
			Type:      string(commonModel.NotificationInboxType),
			Read:      false,
			ReadCount: 0,
			ReadAt:    0,
			Meta:      string(meta),
			CreatedAt: time.Now().UTC().Unix(),
		})
		if err != nil {
			return err
		}
	}

	return id.keyvalueRepo.AddOrUpdateKeyValue(ctx, commonModel.ReleaseVersionKey, latestVersion)
}

func (id *InboxDispatcher) handleInboxClear(ctx context.Context) error {
	const (
		minReadCount      = 2
		readExpireSeconds = 5 * 24 * 60 * 60
	)
	readBefore := time.Now().UTC().Unix() - readExpireSeconds
	expiredReadInboxIDs, err := id.inboxRepo.GetExpiredReadInboxIDs(ctx, minReadCount, readBefore)
	if err != nil {
		return err
	}
	return id.inboxRepo.ClearReadInboxByIds(ctx, expiredReadInboxIDs)
}
