package service

import (
	"context"
	"errors"

	"github.com/lin-snow/ech0/internal/config"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
	jsonUtil "github.com/lin-snow/ech0/internal/util/json"
	"github.com/lin-snow/ech0/pkg/viewer"
)

// GetCommentSetting 获取评论设置
func (settingService *SettingService) GetCommentSetting(setting *model.CommentSetting) error {
	return settingService.transactor.Run(context.Background(), func(ctx context.Context) error {
		commentSetting, err := settingService.keyvalueRepository.GetKeyValue(
			ctx,
			commonModel.CommentSettingKey,
		)
		if err != nil {
			// 数据库中不存在数据时写入新模型默认值
			defaultSetting := settingService.commentRegistry.defaultCommentSetting()
			defaultSetting.EnableComment = config.Config().Comment.EnableComment
			defaultSetting.Provider = config.Config().Comment.Provider
			if err := settingService.commentRegistry.validateProvider(defaultSetting.Provider); err != nil {
				defaultSetting.Provider = string(commonModel.TWIKOO)
			}
			*setting = defaultSetting

			// 序列化为 JSON
			settingToJSON, err := jsonUtil.JSONMarshal(setting)
			if err != nil {
				return err
			}
			if err := settingService.keyvalueRepository.AddKeyValue(ctx, commonModel.CommentSettingKey, string(settingToJSON)); err != nil {
				return err
			}

			return nil
		}

		if err := jsonUtil.JSONUnmarshal([]byte(commentSetting), setting); err != nil {
			return err
		}
		if setting.Providers == nil {
			setting.Providers = settingService.commentRegistry.defaultCommentSetting().Providers
		}
		if err := settingService.commentRegistry.validateProvider(setting.Provider); err != nil {
			setting.Provider = string(commonModel.TWIKOO)
		}

		return nil
	})
}

// UpdateCommentSetting 更新评论设置
func (settingService *SettingService) UpdateCommentSetting(
	ctx context.Context,
	newSetting *model.CommentSettingDto,
) error {
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := settingService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	return settingService.transactor.Run(ctx, func(ctx context.Context) error {
		if err := settingService.commentRegistry.normalizeAndValidate(newSetting); err != nil {
			return err
		}

		commentSetting := &model.CommentSetting{
			EnableComment: newSetting.EnableComment,
			Provider:      newSetting.Provider,
			Providers:     newSetting.Providers,
		}

		// 序列化为 JSON
		settingToJSON, err := jsonUtil.JSONMarshal(commentSetting)
		if err != nil {
			return err
		}

		if err := settingService.keyvalueRepository.AddOrUpdateKeyValue(ctx, commonModel.CommentSettingKey, string(settingToJSON)); err != nil {
			return err
		}

		return nil
	})
}

func (settingService *SettingService) GetCommentProviderMeta() model.CommentProviderMetaResponse {
	return settingService.commentRegistry.providerMeta()
}
