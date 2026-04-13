package service

import (
	"context"
	"errors"
	"time"

	contracts "github.com/lin-snow/ech0/internal/event/contracts"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
	webhookModel "github.com/lin-snow/ech0/internal/model/webhook"
	httpUtil "github.com/lin-snow/ech0/internal/util/http"
	webhookclient "github.com/lin-snow/ech0/internal/webhook/infra/httpclient"
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
	if err := validateWebhookURL(newWebhook.URL); err != nil {
		return err
	}

	// 保存到数据库
	webhook := &webhookModel.Webhook{
		Name:     newWebhook.Name,
		URL:      newWebhook.URL,
		Secret:   newWebhook.Secret,
		IsActive: newWebhook.IsActive,
	}

	return settingService.transactor.Run(ctx, func(ctx context.Context) error {
		return settingService.webhookRepository.UpdateWebhookByID(ctx, id, webhook)
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
	if err := validateWebhookURL(newWebhook.URL); err != nil {
		return err
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

// TestWebhook 测试单个 Webhook
func (settingService *SettingService) TestWebhook(ctx context.Context, id string) error {
	// 鉴权
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := settingService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	webhook, err := settingService.webhookRepository.GetWebhookByID(ctx, id)
	if err != nil {
		return err
	}
	if err := validateWebhookURL(webhook.URL); err != nil {
		return err
	}

	payload := map[string]any{
		"message": "webhook connectivity test from ech0",
		"webhook": webhook.Name,
		"time":    time.Now().UTC().Format(time.RFC3339),
	}
	obs, err := contracts.NewWebhookObservation("webhook.test", payload, map[string]string{
		"source": "setting.test",
	})
	if err != nil {
		return err
	}

	client := webhookclient.NewSafeHTTPClient(5 * time.Second)
	triggerAt := time.Now().UTC().Unix()
	sendErr := webhookclient.SendWithRetry(client, webhook, obs, 2, 300*time.Millisecond)
	status := "success"
	if sendErr != nil {
		status = "failed"
	}
	_ = settingService.webhookRepository.UpdateWebhookDeliveryStatus(ctx, webhook.ID, status, triggerAt)
	return sendErr
}

func validateWebhookURL(rawURL string) error {
	if err := httpUtil.ValidatePublicHTTPURL(rawURL); err != nil {
		return errors.New(commonModel.INVALID_WEBHOOK_URL)
	}
	return nil
}
