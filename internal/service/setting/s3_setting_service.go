// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"errors"
	"strings"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
	coreSetting "github.com/lin-snow/ech0/internal/setting"
	urlUtil "github.com/lin-snow/ech0/internal/util/url"
	"github.com/lin-snow/ech0/pkg/viewer"
)

// GetS3Setting 获取 S3 存储设置。缺省值由 setting 引擎处理；脱敏（非管理员屏蔽敏感字段）
// 属请求态逻辑，留在此处。
func (settingService *SettingService) GetS3Setting(ctx context.Context, setting *model.S3Setting) error {
	userid := viewer.MustFromContext(ctx).UserID()
	v, err := coreSetting.Get(ctx, settingService.durableKV, coreSetting.S3)
	if err != nil {
		return err
	}
	*setting = v

	// 未登录直接脱敏。
	if userid == "" {
		maskS3Secrets(setting)
		return nil
	}
	user, err := settingService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		maskS3Secrets(setting)
	}
	return nil
}

func maskS3Secrets(setting *model.S3Setting) {
	setting.AccessKey = "******"
	setting.SecretKey = "******"
	setting.BucketName = "******"
	setting.Endpoint = "******"
}

// UpdateS3Setting 更新 S3 存储设置
func (settingService *SettingService) UpdateS3Setting(
	ctx context.Context,
	newSetting *model.S3SettingDto,
) error {
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := settingService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	oldRaw, _ := settingService.durableKV.Get(ctx, commonModel.S3SettingKey)
	var appliedSetting *model.S3Setting

	err = settingService.transactor.Run(ctx, func(ctx context.Context) error {
		// 检查endpoint是否为http(s)动态改变USE SSL
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(newSetting.Endpoint)), "https://") {
			newSetting.UseSSL = true
		} else if strings.HasPrefix(strings.ToLower(strings.TrimSpace(newSetting.Endpoint)), "http://") {
			newSetting.UseSSL = false
		}

		// 去除Endpoint的协议头（http://或https://）
		endpoint := strings.TrimSpace(newSetting.Endpoint)
		endpoint = strings.TrimPrefix(endpoint, "http://")
		endpoint = strings.TrimPrefix(endpoint, "https://")
		newSetting.Endpoint = endpoint

		cdnURL := strings.TrimSpace(newSetting.CDNURL)
		if cdnURL != "" {
			cdnURL = strings.TrimRight(cdnURL, "/")
		}

		s3Setting := &model.S3Setting{
			Enable:     newSetting.Enable,
			Provider:   newSetting.Provider,
			Endpoint:   urlUtil.TrimURL(newSetting.Endpoint),
			AccessKey:  newSetting.AccessKey,
			SecretKey:  newSetting.SecretKey,
			BucketName: newSetting.BucketName,
			Region:     strings.TrimSpace(newSetting.Region),
			UseSSL:     newSetting.UseSSL,
			CDNURL:     cdnURL,
			PathPrefix: urlUtil.TrimURL(newSetting.PathPrefix),
			PublicRead: newSetting.PublicRead,
		}

		// 配置检查
		switch s3Setting.Provider {
		case string(commonModel.R2):
			if s3Setting.Region == "" {
				s3Setting.Region = "auto"
			}
			s3Setting.UseSSL = true
		case string(commonModel.AWS):
			if s3Setting.Region == "" {
				s3Setting.Region = "us-east-1"
			}
		case string(commonModel.MINIO):
			if s3Setting.Region == "" {
				s3Setting.Region = "us-east-1"
			}
		case string(commonModel.OTHER):
			// 其他 S3 兼容厂商（Backblaze、Wasabi、Ceph 等）
			if s3Setting.Region == "" {
				s3Setting.Region = "auto"
			}
		default:
		}

		if err := coreSetting.Set(ctx, settingService.durableKV, coreSetting.S3, *s3Setting); err != nil {
			return err
		}

		appliedSetting = s3Setting
		return nil
	})
	if err != nil {
		return err
	}

	if settingService.storageManager != nil && appliedSetting != nil {
		if err := settingService.storageManager.ApplyS3Setting(*appliedSetting); err != nil {
			_ = settingService.transactor.Run(context.Background(), func(ctx context.Context) error {
				if strings.TrimSpace(oldRaw) == "" {
					return settingService.durableKV.Delete(ctx, commonModel.S3SettingKey)
				}
				return settingService.durableKV.Set(ctx, commonModel.S3SettingKey, oldRaw)
			})
			_ = settingService.storageManager.ReloadFromConfigAndDB(context.Background())
			return err
		}
	}

	return nil
}
