package service

import (
	"context"
	"errors"
	"strings"

	"github.com/lin-snow/ech0/internal/config"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
	httpUtil "github.com/lin-snow/ech0/internal/util/http"
	jsonUtil "github.com/lin-snow/ech0/internal/util/json"
	"github.com/lin-snow/ech0/pkg/viewer"
)

// GetS3Setting 获取 S3 存储设置
func (settingService *SettingService) GetS3Setting(ctx context.Context, setting *model.S3Setting) error {
	userid := viewer.MustFromContext(ctx).UserID()
	return settingService.transactor.Run(ctx, func(ctx context.Context) error {
		s3Setting, err := settingService.keyvalueRepository.GetKeyValue(ctx, commonModel.S3SettingKey)
		if err != nil {
			// 数据库缺失时回退到 config 默认值
			cfg := config.Config().Storage
			setting.Enable = cfg.ObjectEnabled
			setting.Provider = strings.TrimSpace(cfg.Provider)
			setting.Endpoint = strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(cfg.Endpoint), "http://"), "https://")
			setting.AccessKey = cfg.AccessKey
			setting.SecretKey = cfg.SecretKey
			setting.BucketName = cfg.BucketName
			setting.Region = strings.TrimSpace(cfg.Region)
			setting.UseSSL = cfg.UseSSL
			setting.CDNURL = strings.TrimRight(strings.TrimSpace(cfg.CDNURL), "/")
			setting.PathPrefix = strings.Trim(strings.TrimSpace(cfg.PathPrefix), "/")
			setting.PublicRead = true

			// 序列化为 JSON
			settingToJSON, err := jsonUtil.JSONMarshal(setting)
			if err != nil {
				return err
			}
			if err := settingService.keyvalueRepository.AddKeyValue(ctx, commonModel.S3SettingKey, string(settingToJSON)); err != nil {
				return err
			}

			return nil
		}

		if err := jsonUtil.JSONUnmarshal([]byte(s3Setting), setting); err != nil {
			return err
		}

		// 如果用户未登录且不为管理员,则屏蔽 S3 设置的敏感信息
		if userid == "" {
			setting.AccessKey = "******"
			setting.SecretKey = "******"
			setting.BucketName = "******"
			setting.Endpoint = "******"
		} else {
			user, err := settingService.commonService.CommonGetUserByUserId(ctx, userid)
			if err != nil {
				return err
			}
			if !user.IsAdmin {
				setting.AccessKey = "******"
				setting.SecretKey = "******"
				setting.BucketName = "******"
				setting.Endpoint = "******"
			}
		}

		return nil
	})
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

	oldRaw, _ := settingService.keyvalueRepository.GetKeyValue(ctx, commonModel.S3SettingKey)
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
			Endpoint:   httpUtil.TrimURL(newSetting.Endpoint),
			AccessKey:  newSetting.AccessKey,
			SecretKey:  newSetting.SecretKey,
			BucketName: newSetting.BucketName,
			Region:     strings.TrimSpace(newSetting.Region),
			UseSSL:     newSetting.UseSSL,
			CDNURL:     cdnURL,
			PathPrefix: httpUtil.TrimURL(newSetting.PathPrefix),
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

		// 序列化为 JSON
		settingToJSON, err := jsonUtil.JSONMarshal(s3Setting)
		if err != nil {
			return err
		}

		if err := settingService.keyvalueRepository.AddOrUpdateKeyValue(ctx, commonModel.S3SettingKey, string(settingToJSON)); err != nil {
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
					return settingService.keyvalueRepository.DeleteKeyValue(ctx, commonModel.S3SettingKey)
				}
				return settingService.keyvalueRepository.AddOrUpdateKeyValue(ctx, commonModel.S3SettingKey, oldRaw)
			})
			_ = settingService.storageManager.ReloadFromConfigAndDB(context.Background())
			return err
		}
	}

	return nil
}
