// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	contracts "github.com/lin-snow/ech0/internal/event/contracts"
	publisher "github.com/lin-snow/ech0/internal/event/publisher"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/echo"
	"github.com/lin-snow/ech0/internal/storage"
	"github.com/lin-snow/ech0/internal/transaction"
	httpUtil "github.com/lin-snow/ech0/internal/util/http"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"github.com/lin-snow/ech0/pkg/viewer"
	"go.uber.org/zap"
)

type EchoService struct {
	transactor     transaction.Transactor
	commonService  CommonService
	fileService    FileService
	echoRepository Repository
	publisher      *publisher.Publisher
}

func NewEchoService(
	tx transaction.Transactor,
	commonService CommonService,
	fileService FileService,
	echoRepository Repository,
	publisher *publisher.Publisher,
) *EchoService {
	return &EchoService{
		transactor:     tx,
		commonService:  commonService,
		fileService:    fileService,
		echoRepository: echoRepository,
		publisher:      publisher,
	}
}

func (echoService *EchoService) PostEcho(ctx context.Context, newEcho *model.Echo) error {
	userid := viewer.MustFromContext(ctx).UserID()
	newEcho.UserID = userid

	user, err := echoService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return err
	}

	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	layout := strings.TrimSpace(newEcho.Layout)
	if layout == "" || (layout != model.LayoutWaterfall &&
		layout != model.LayoutGrid &&
		layout != model.LayoutHorizontal &&
		layout != model.LayoutCarousel &&
		layout != model.LayoutStack) {
		newEcho.Layout = model.LayoutWaterfall
	}

	normalizedExt, err := normalizeEchoExtension(newEcho.Extension)
	if err != nil {
		return err
	}
	newEcho.Extension = normalizedExt

	newEcho.Username = user.Username

	if isEchoEmpty(newEcho) {
		return errors.New(commonModel.ECHO_CAN_NOT_BE_EMPTY)
	}

	if err := echoService.transactor.Run(ctx, func(txCtx context.Context) error {
		if err := echoService.ProcessEchoTags(txCtx, newEcho); err != nil {
			return err
		}
		return echoService.echoRepository.CreateEcho(txCtx, newEcho)
	}); err != nil {
		return err
	}

	echoService.echoRepository.InvalidateEchoCaches()

	savedEcho, fetchErr := echoService.echoRepository.GetEchosById(ctx, newEcho.ID)
	if fetchErr != nil {
		return fetchErr
	}
	if savedEcho != nil {
		if pubErr := echoService.publisher.EchoCreated(
			context.Background(),
			contracts.EchoCreatedEvent{Echo: *savedEcho, User: user},
		); pubErr != nil {
			logUtil.GetLogger().Error("publish echo created event failed", zap.Error(pubErr))
		}
	}
	if err := echoService.fileService.ConfirmTempFiles(ctx, collectEchoFileIDs(newEcho)); err != nil {
		logUtil.GetLogger().Warn("confirm temp files after post echo failed", zap.Error(err))
	}

	return nil
}

// GetEchosByPage Deprecated: use QueryEchos instead. Kept for backward compatibility.
func (echoService *EchoService) GetEchosByPage(
	ctx context.Context,
	pageQueryDto commonModel.PageQueryDto,
) (commonModel.PageQueryResult[[]model.Echo], error) {
	return echoService.QueryEchos(ctx, commonModel.EchoQueryDto{
		Page:     pageQueryDto.Page,
		PageSize: pageQueryDto.PageSize,
		Search:   pageQueryDto.Search,
	})
}

func (echoService *EchoService) DeleteEchoById(ctx context.Context, id string) error {
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := echoService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	type deletableFileRef struct {
		key         string
		storageType string
	}
	var deletableFiles []deletableFileRef
	if err := echoService.transactor.Run(ctx, func(txCtx context.Context) error {
		echo, err := echoService.echoRepository.GetEchosById(txCtx, id)
		if err != nil {
			return err
		}
		if echo == nil {
			return errors.New(commonModel.ECHO_NOT_FOUND)
		}

		for _, ef := range echo.EchoFiles {
			if ef.File.Key != "" && storage.NormalizeStorageType(ef.File.StorageType) != storage.StorageTypeExternal {
				deletableFiles = append(deletableFiles, deletableFileRef{
					key:         ef.File.Key,
					storageType: ef.File.StorageType,
				})
			}
			if ef.File.ID != "" {
				if err := echoService.fileService.DeleteFileRecord(txCtx, ef.File.ID); err != nil {
					return err
				}
			}
		}

		return echoService.echoRepository.DeleteEchoById(txCtx, id)
	}); err != nil {
		return err
	}

	echoService.echoRepository.InvalidateEchoCaches(id)

	if pubErr := echoService.publisher.EchoDeleted(
		context.Background(),
		contracts.EchoDeletedEvent{Echo: model.Echo{ID: id}, User: user},
	); pubErr != nil {
		logUtil.GetLogger().Error("publish echo deleted event failed", zap.Error(pubErr))
	}

	for _, file := range deletableFiles {
		_ = echoService.fileService.DeleteStoredFile(file.storageType, file.key)
	}

	return nil
}

func (echoService *EchoService) GetTodayEchos(ctx context.Context, timezone string) ([]model.Echo, error) {
	userid := viewer.MustFromContext(ctx).UserID()
	showPrivate := false
	if userid != "" {
		user, err := echoService.commonService.CommonGetUserByUserId(ctx, userid)
		if err != nil {
			return nil, err
		}
		showPrivate = user.IsAdmin
	}

	todayEchos := echoService.echoRepository.GetTodayEchos(showPrivate, timezone)
	return todayEchos, nil
}

func (echoService *EchoService) GetHotEchos(ctx context.Context, limit int) ([]model.Echo, error) {
	userid := viewer.MustFromContext(ctx).UserID()
	showPrivate := false
	if userid != "" {
		user, err := echoService.commonService.CommonGetUserByUserId(ctx, userid)
		if err != nil {
			return nil, err
		}
		showPrivate = user.IsAdmin
	}
	return echoService.echoRepository.GetHotEchos(limit, showPrivate)
}

func (echoService *EchoService) UpdateEcho(ctx context.Context, echo *model.Echo) error {
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := echoService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	layout := strings.TrimSpace(echo.Layout)
	if layout == "" || (layout != model.LayoutWaterfall &&
		layout != model.LayoutGrid &&
		layout != model.LayoutHorizontal &&
		layout != model.LayoutCarousel &&
		layout != model.LayoutStack) {
		echo.Layout = model.LayoutWaterfall
	}

	normalizedExt, err := normalizeEchoExtension(echo.Extension)
	if err != nil {
		return err
	}
	echo.Extension = normalizedExt

	for i := range echo.EchoFiles {
		echo.EchoFiles[i].EchoID = echo.ID
	}

	if isEchoEmpty(echo) {
		return errors.New(commonModel.ECHO_CAN_NOT_BE_EMPTY)
	}

	if err := echoService.transactor.Run(ctx, func(txCtx context.Context) error {
		if err := echoService.ProcessEchoTags(txCtx, echo); err != nil {
			return err
		}
		return echoService.echoRepository.UpdateEcho(txCtx, echo)
	}); err != nil {
		return err
	}

	echoService.echoRepository.InvalidateEchoCaches(echo.ID)

	if pubErr := echoService.publisher.EchoUpdated(
		context.Background(),
		contracts.EchoUpdatedEvent{Echo: *echo, User: user},
	); pubErr != nil {
		logUtil.GetLogger().Error("publish echo updated event failed", zap.Error(pubErr))
	}
	if err := echoService.fileService.ConfirmTempFiles(ctx, collectEchoFileIDs(echo)); err != nil {
		logUtil.GetLogger().Warn("confirm temp files after update echo failed", zap.Error(err))
	}

	return nil
}

func (echoService *EchoService) LikeEcho(ctx context.Context, id string) error {
	if err := echoService.transactor.Run(ctx, func(txCtx context.Context) error {
		return echoService.echoRepository.LikeEcho(txCtx, id)
	}); err != nil {
		return err
	}
	echoService.echoRepository.InvalidateEchoCaches(id)
	return nil
}

func (echoService *EchoService) GetEchoById(ctx context.Context, id string) (*model.Echo, error) {
	userId := viewer.MustFromContext(ctx).UserID()
	echo, err := echoService.echoRepository.GetEchosById(ctx, id)
	if err != nil {
		return nil, err
	}
	if echo == nil {
		return nil, errors.New(commonModel.ECHO_NOT_FOUND)
	}

	if userId == "" {
		if echo.Private {
			return nil, errors.New(commonModel.NO_PERMISSION_DENIED)
		}
	} else {
		user, err := echoService.commonService.CommonGetUserByUserId(ctx, userId)
		if err != nil {
			return nil, err
		}
		if echo.Private && !user.IsAdmin {
			return nil, errors.New(commonModel.NO_PERMISSION_DENIED)
		}
	}

	return echo, nil
}

func (echoService *EchoService) GetAllTags() ([]model.Tag, error) {
	return echoService.echoRepository.GetAllTags()
}

func (echoService *EchoService) DeleteTag(ctx context.Context, id string) error {
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := echoService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	return echoService.transactor.Run(ctx, func(txCtx context.Context) error {
		return echoService.echoRepository.DeleteTagById(txCtx, id)
	})
}

func (echoService *EchoService) ProcessEchoTags(ctx context.Context, echo *model.Echo) error {
	var processedTags []model.Tag

	var names []string
	for _, tag := range echo.Tags {
		name := strings.TrimSpace(strings.TrimPrefix(tag.Name, "#"))
		if name != "" {
			names = append(names, name)
		}
	}

	existingTags, err := echoService.echoRepository.GetTagsByNames(ctx, names)
	if err != nil {
		return err
	}

	existingMap := make(map[string]*model.Tag)
	for _, t := range existingTags {
		existingMap[t.Name] = t
	}

	for _, name := range names {
		if existing, ok := existingMap[name]; ok {
			if err := echoService.echoRepository.IncrementTagUsageCount(ctx, existing.ID); err != nil {
				return err
			}
			processedTags = append(processedTags, *existing)
		} else {
			newTag := model.Tag{Name: name, UsageCount: 1}
			if err := echoService.echoRepository.CreateTag(ctx, &newTag); err != nil {
				return err
			}
			processedTags = append(processedTags, newTag)
		}
	}

	echo.Tags = processedTags
	return nil
}

// GetEchosByTagId Deprecated: use QueryEchos instead. Kept for backward compatibility.
func (echoService *EchoService) GetEchosByTagId(
	ctx context.Context,
	tagId string,
	pageQueryDto commonModel.PageQueryDto,
) (commonModel.PageQueryResult[[]model.Echo], error) {
	return echoService.QueryEchos(ctx, commonModel.EchoQueryDto{
		Page:     pageQueryDto.Page,
		PageSize: pageQueryDto.PageSize,
		Search:   pageQueryDto.Search,
		TagIDs:   []string{tagId},
	})
}

func (echoService *EchoService) QueryEchos(
	ctx context.Context,
	queryDto commonModel.EchoQueryDto,
) (commonModel.PageQueryResult[[]model.Echo], error) {
	if queryDto.Page < 1 {
		queryDto.Page = 1
	}
	if queryDto.PageSize < 1 || queryDto.PageSize > 100 {
		queryDto.PageSize = 10
	}
	queryDto.Search = strings.TrimSpace(queryDto.Search)

	if queryDto.SortBy == "" {
		queryDto.SortBy = "created_at"
	}
	if queryDto.SortOrder == "" {
		queryDto.SortOrder = "desc"
	}

	userId := viewer.MustFromContext(ctx).UserID()
	showPrivate := false
	if userId != "" {
		user, err := echoService.commonService.CommonGetUserByUserId(ctx, userId)
		if err != nil {
			return commonModel.PageQueryResult[[]model.Echo]{}, err
		}
		showPrivate = user.IsAdmin
	}

	echos, total, err := echoService.echoRepository.QueryEchos(queryDto, showPrivate)
	if err != nil {
		return commonModel.PageQueryResult[[]model.Echo]{}, err
	}

	return commonModel.PageQueryResult[[]model.Echo]{
		Items: echos,
		Total: total,
	}, nil
}

func normalizeEchoExtension(ext *model.EchoExtension) (*model.EchoExtension, error) {
	if ext == nil {
		return nil, nil
	}

	extType := strings.TrimSpace(ext.Type)
	if extType == "" {
		return nil, nil
	}
	ext.Type = extType
	if ext.Payload == nil {
		return nil, fmt.Errorf("extension payload is required")
	}

	switch ext.Type {
	case model.Extension_MUSIC:
		url := strings.TrimSpace(getPayloadString(ext.Payload, "url"))
		if url == "" {
			return nil, fmt.Errorf("extension payload.url is required for MUSIC")
		}
		ext.Payload = map[string]interface{}{"url": httpUtil.TrimURL(url)}
	case model.Extension_VIDEO:
		videoID := strings.TrimSpace(getPayloadString(ext.Payload, "videoId"))
		if videoID == "" {
			return nil, fmt.Errorf("extension payload.videoId is required for VIDEO")
		}
		ext.Payload = map[string]interface{}{"videoId": videoID}
	case model.Extension_GITHUBPROJ:
		repoURL := strings.TrimSpace(getPayloadString(ext.Payload, "repoUrl"))
		if repoURL == "" {
			return nil, fmt.Errorf("extension payload.repoUrl is required for GITHUBPROJ")
		}
		ext.Payload = map[string]interface{}{"repoUrl": httpUtil.TrimURL(repoURL)}
	case model.Extension_WEBSITE:
		title := strings.TrimSpace(getPayloadString(ext.Payload, "title"))
		site := strings.TrimSpace(getPayloadString(ext.Payload, "site"))
		if title == "" || site == "" {
			return nil, fmt.Errorf("extension payload.title and payload.site are required for WEBSITE")
		}
		ext.Payload = map[string]interface{}{
			"title": title,
			"site":  httpUtil.TrimURL(site),
		}
	case model.Extension_LOCATION:
		lat, okLat := getPayloadFloat(ext.Payload, "latitude")
		lng, okLng := getPayloadFloat(ext.Payload, "longitude")
		if !okLat || !okLng {
			return nil, fmt.Errorf("extension payload.latitude and payload.longitude are required for LOCATION")
		}
		if lat < -90 || lat > 90 || lng < -180 || lng > 180 {
			return nil, fmt.Errorf("extension payload.latitude/longitude out of range for LOCATION")
		}
		placeholder := strings.TrimSpace(getPayloadString(ext.Payload, "placeholder"))
		if placeholder == "" {
			return nil, fmt.Errorf("extension payload.placeholder is required for LOCATION")
		}
		ext.Payload = map[string]interface{}{
			"latitude":    lat,
			"longitude":   lng,
			"placeholder": placeholder,
		}
	default:
		return nil, fmt.Errorf("unsupported extension type: %s", ext.Type)
	}

	return ext, nil
}

func getPayloadString(payload map[string]interface{}, key string) string {
	raw, ok := payload[key]
	if !ok || raw == nil {
		return ""
	}
	value, ok := raw.(string)
	if !ok {
		return ""
	}
	return value
}

// getPayloadFloat extracts a float64 from a JSON-decoded payload.
// JSON numbers land as float64 by default; accept int and string-encoded numbers as fallbacks.
func getPayloadFloat(payload map[string]interface{}, key string) (float64, bool) {
	raw, ok := payload[key]
	if !ok || raw == nil {
		return 0, false
	}
	switch v := raw.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case json.Number:
		if f, err := v.Float64(); err == nil {
			return f, true
		}
	case string:
		if f, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

func isEchoEmpty(echo *model.Echo) bool {
	if echo == nil {
		return true
	}
	content := strings.TrimSpace(echo.Content)
	return content == "" && len(echo.EchoFiles) == 0 && echo.Extension == nil
}

func collectEchoFileIDs(echo *model.Echo) []string {
	if echo == nil || len(echo.EchoFiles) == 0 {
		return nil
	}
	ids := make([]string, 0, len(echo.EchoFiles))
	for _, ef := range echo.EchoFiles {
		if strings.TrimSpace(ef.FileID) != "" {
			ids = append(ids, ef.FileID)
			continue
		}
		if strings.TrimSpace(ef.File.ID) != "" {
			ids = append(ids, ef.File.ID)
		}
	}
	return ids
}
