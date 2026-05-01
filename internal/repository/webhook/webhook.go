// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository

import (
	"context"
	"errors"

	model "github.com/lin-snow/ech0/internal/model/webhook"
	settingService "github.com/lin-snow/ech0/internal/service/setting"
	"github.com/lin-snow/ech0/internal/transaction"
	"gorm.io/gorm"
)

type WebhookRepository struct {
	db func() *gorm.DB
}

var _ settingService.WebhookRepository = (*WebhookRepository)(nil)

func NewWebhookRepository(dbProvider func() *gorm.DB) *WebhookRepository {
	return &WebhookRepository{
		db: dbProvider,
	}
}

func (webhookRepository *WebhookRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := transaction.TxFromContext(ctx); ok {
		return tx
	}
	return webhookRepository.db()
}

// CreateWebhook 创建一个webhook
func (webhookRepository *WebhookRepository) CreateWebhook(
	ctx context.Context,
	webhook *model.Webhook,
) error {
	if err := webhookRepository.getDB(ctx).Create(webhook).Error; err != nil {
		return err
	}

	return nil
}

// UpdateWebhookByID 根据ID更新 webhook
func (webhookRepository *WebhookRepository) UpdateWebhookByID(
	ctx context.Context,
	id string,
	webhook *model.Webhook,
) error {
	tx := webhookRepository.getDB(ctx).
		Model(&model.Webhook{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"name":      webhook.Name,
			"url":       webhook.URL,
			"secret":    webhook.Secret,
			"is_active": webhook.IsActive,
		})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("webhook not found")
	}
	return nil
}

// GetAllWebhooks 获取所有webhooks
func (webhookRepository *WebhookRepository) GetAllWebhooks(ctx context.Context) ([]model.Webhook, error) {
	var webhooks []model.Webhook
	if err := webhookRepository.getDB(ctx).Find(&webhooks).Error; err != nil {
		return nil, err
	}

	return webhooks, nil
}

// GetWebhookByID 根据 ID 获取 webhook
func (webhookRepository *WebhookRepository) GetWebhookByID(ctx context.Context, id string) (*model.Webhook, error) {
	var webhook model.Webhook
	err := webhookRepository.getDB(ctx).Where("id = ?", id).First(&webhook).Error
	if err != nil {
		return nil, err
	}
	return &webhook, nil
}

// DeleteWebhookByID 根据ID删除webhook
func (webhookRepository *WebhookRepository) DeleteWebhookByID(ctx context.Context, id string) error {
	if err := webhookRepository.getDB(ctx).Where("id = ?", id).Delete(&model.Webhook{}).Error; err != nil {
		return err
	}

	return nil
}

// ListActiveWebhooks 列出所有激活的 webhook
func (webhookRepository *WebhookRepository) ListActiveWebhooks(ctx context.Context) ([]model.Webhook, error) {
	var webhooks []model.Webhook
	if err := webhookRepository.getDB(ctx).Where("is_active = ?", true).Find(&webhooks).Error; err != nil {
		return nil, err
	}

	return webhooks, nil
}

// UpdateWebhookDeliveryStatus 更新 webhook 最近投递状态
func (webhookRepository *WebhookRepository) UpdateWebhookDeliveryStatus(
	ctx context.Context,
	id string,
	status string,
	lastTrigger int64,
) error {
	tx := webhookRepository.getDB(ctx).
		Model(&model.Webhook{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"last_status":  status,
			"last_trigger": lastTrigger,
		})
	return tx.Error
}
