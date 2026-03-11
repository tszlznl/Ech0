package service

import (
	"context"
	"errors"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
	webhookModel "github.com/lin-snow/ech0/internal/model/webhook"
	httpUtil "github.com/lin-snow/ech0/internal/util/http"
	"github.com/lin-snow/ech0/pkg/viewer"
)

// GetAllWebhooks 获取所有 Webhook
func (settingService *SettingService) GetAllWebhooks(ctx context.Context) ([]webhookModel.Webhook, error) {
	// 鉴权
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := settingService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return nil, err
	}
	if !user.IsAdmin {
		return nil, errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	webhooks, err := settingService.webhookRepository.GetAllWebhooks(ctx)
	if err != nil {
		return nil, err
	}

	return webhooks, nil
}

// DeleteWebhook 删除 Webhook
func (settingService *SettingService) DeleteWebhook(ctx context.Context, id string) error {
	// 鉴权
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := settingService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	return settingService.transactor.Run(ctx, func(txCtx context.Context) error {
		return settingService.webhookRepository.DeleteWebhookByID(txCtx, id)
	})
}

// UpdateWebhook 更新 Webhook
func (settingService *SettingService) UpdateWebhook(
	ctx context.Context,
	id string,
	newWebhook *model.WebhookDto,
) error {
	// 鉴权
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := settingService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	// 数据处理
	newWebhook.URL = httpUtil.TrimURL(newWebhook.URL)

	// 检查名称或URL是否为空
	if newWebhook.Name == "" || newWebhook.URL == "" {
		return errors.New(commonModel.WEBHOOK_NAME_OR_URL_CANNOT_BE_EMPTY)
	}

	// 保存到数据库
	webhook := &webhookModel.Webhook{
		ID:       id,
		Name:     newWebhook.Name,
		URL:      newWebhook.URL,
		Secret:   newWebhook.Secret,
		IsActive: newWebhook.IsActive,
	}

	return settingService.transactor.Run(ctx, func(ctx context.Context) error {
		// 先删除再创建，避免部分字段无法更新的问题
		if err := settingService.webhookRepository.DeleteWebhookByID(ctx, webhook.ID); err != nil {
			return err
		}
		return settingService.webhookRepository.CreateWebhook(ctx, webhook)
	})
}

// CreateWebhook 创建 Webhook
func (settingService *SettingService) CreateWebhook(
	ctx context.Context,
	newWebhook *model.WebhookDto,
) error {
	// 鉴权
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := settingService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	// 数据处理
	newWebhook.URL = httpUtil.TrimURL(newWebhook.URL)

	// 检查名称或URL是否为空
	if newWebhook.Name == "" || newWebhook.URL == "" {
		return errors.New(commonModel.WEBHOOK_NAME_OR_URL_CANNOT_BE_EMPTY)
	}

	// 保存到数据库
	webhook := &webhookModel.Webhook{
		Name:     newWebhook.Name,
		URL:      newWebhook.URL,
		Secret:   newWebhook.Secret,
		IsActive: newWebhook.IsActive,
	}

	return settingService.transactor.Run(ctx, func(ctx context.Context) error {
		return settingService.webhookRepository.CreateWebhook(ctx, webhook)
	})
}
