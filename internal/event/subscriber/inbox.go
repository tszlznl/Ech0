package subscriber

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	contracts "github.com/lin-snow/ech0/internal/event/contracts"
	registry "github.com/lin-snow/ech0/internal/event/registry"
	i18nUtil "github.com/lin-snow/ech0/internal/i18n"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	inboxModel "github.com/lin-snow/ech0/internal/model/inbox"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	agentService "github.com/lin-snow/ech0/internal/service/agent"
	githubUtil "github.com/lin-snow/ech0/internal/util/github"
	"golang.org/x/mod/semver"
	"gorm.io/gorm"
)

type InboxStore interface {
	PostInbox(ctx context.Context, inbox *inboxModel.Inbox) error
	GetExpiredReadInboxIDs(ctx context.Context, minReadCount int, readBefore int64) ([]string, error)
	ClearReadInboxByIds(ctx context.Context, inboxIDs []string) error
}

type InboxDispatcher struct {
	inboxRepo    InboxStore
	keyvalueRepo agentService.KeyValueRepository
}

func NewInboxDispatcher(inboxRepo InboxStore, keyvalueRepo agentService.KeyValueRepository) *InboxDispatcher {
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
		systemLocale := id.getSystemDefaultLocale(ctx)
		localizer := i18nUtil.NewLocalizer(systemLocale, "")
		content := i18nUtil.Localize(
			localizer,
			commonModel.MsgKeyInboxNewVersion,
			"有新版本可用，请更新："+latestVersion,
			map[string]any{"LatestVersion": latestVersion},
		)
		meta, _ := json.Marshal(map[string]string{
			"latest_version":  latestVersion,
			"current_version": currentVersion,
			"message_key":     commonModel.MsgKeyInboxNewVersion,
			"locale":          systemLocale,
		})
		err = id.inboxRepo.PostInbox(ctx, &inboxModel.Inbox{
			Source:    string(commonModel.SystemSource),
			Content:   content,
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

func (id *InboxDispatcher) Subscriptions() []registry.Subscription {
	return []registry.Subscription{
		registry.TopicSubscription(
			contracts.TopicEch0UpdateCheck,
			id.HandleEch0UpdateCheck,
			registry.InboxSubscribeOptions()...,
		),
		registry.TopicSubscription(
			contracts.TopicInboxClear,
			id.HandleInboxClear,
			registry.InboxSubscribeOptions()...,
		),
	}
}

func (id *InboxDispatcher) getSystemDefaultLocale(ctx context.Context) string {
	systemSettingRaw, err := id.keyvalueRepo.GetKeyValue(ctx, commonModel.SystemSettingsKey)
	if err != nil || strings.TrimSpace(systemSettingRaw) == "" {
		return string(commonModel.DefaultLocale)
	}
	var systemSetting settingModel.SystemSetting
	if json.Unmarshal([]byte(systemSettingRaw), &systemSetting) != nil {
		return string(commonModel.DefaultLocale)
	}
	locale := strings.TrimSpace(systemSetting.DefaultLocale)
	if locale == "" {
		return string(commonModel.DefaultLocale)
	}
	return i18nUtil.ResolveLocale(locale)
}
