package service

import (
	"context"
	"errors"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
	httpUtil "github.com/lin-snow/ech0/internal/util/http"
	jsonUtil "github.com/lin-snow/ech0/internal/util/json"
	"github.com/lin-snow/ech0/pkg/viewer"
)

// GetAgentInfo 获取 Agent 信息
func (settingService *SettingService) GetAgentInfo(setting *model.AgentSetting) error {
	return settingService.transactor.Run(context.Background(), func(ctx context.Context) error {
		agentSetting, err := settingService.keyvalueRepository.GetKeyValue(
			ctx,
			commonModel.AgentSettingKey,
		)
		if err != nil {
			// 数据库中不存在数据，返回默认值
			setting.Enable = false
			setting.Provider = string(commonModel.OpenAI)
			setting.Model = ""
			setting.ApiKey = ""
			setting.Prompt = ""
			setting.BaseURL = ""

			// 序列化为 JSON
			settingToJSON, err := jsonUtil.JSONMarshal(setting)
			if err != nil {
				return err
			}
			if err := settingService.keyvalueRepository.AddKeyValue(ctx, commonModel.AgentSettingKey, string(settingToJSON)); err != nil {
				return err
			}
			return nil
		}

		if err := jsonUtil.JSONUnmarshal([]byte(agentSetting), setting); err != nil {
			return err
		}

		return nil
	})
}

// GetAgentSettings 获取 Agent 设置
func (settingService *SettingService) GetAgentSettings(
	ctx context.Context,
	setting *model.AgentSetting,
) error {
	// 检查用户权限
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := settingService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	return settingService.transactor.Run(ctx, func(ctx context.Context) error {
		agentSetting, err := settingService.keyvalueRepository.GetKeyValue(
			ctx,
			commonModel.AgentSettingKey,
		)
		if err != nil {
			// 数据库中不存在数据，返回默认值
			setting.Enable = false
			setting.Provider = string(commonModel.OpenAI)
			setting.Model = ""
			setting.ApiKey = ""
			setting.Prompt = ""
			setting.BaseURL = ""

			// 序列化为 JSON
			settingToJSON, err := jsonUtil.JSONMarshal(setting)
			if err != nil {
				return err
			}
			if err := settingService.keyvalueRepository.AddKeyValue(ctx, commonModel.AgentSettingKey, string(settingToJSON)); err != nil {
				return err
			}

			return nil
		}

		if err := jsonUtil.JSONUnmarshal([]byte(agentSetting), setting); err != nil {
			return err
		}

		return nil
	})
}

// UpdateAgentSettings 更新 Agent 设置
func (settingService *SettingService) UpdateAgentSettings(
	ctx context.Context,
	newSetting *model.AgentSettingDto,
) error {
	// 检查用户权限
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := settingService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	if newSetting.Provider != string(commonModel.OpenAI) &&
		newSetting.Provider != string(commonModel.DeepSeek) &&
		newSetting.Provider != string(commonModel.Anthropic) &&
		newSetting.Provider != string(commonModel.Gemini) &&
		newSetting.Provider != string(commonModel.Qwen) &&
		newSetting.Provider != string(commonModel.Ollama) &&
		newSetting.Provider != string(commonModel.Custom) {
		newSetting.Provider = string(commonModel.Custom) // 如果提供商不在列表中，默认为 Custom
	}

	setting := &model.AgentSetting{
		Enable:   newSetting.Enable,
		Provider: newSetting.Provider,
		Model:    newSetting.Model,
		ApiKey:   newSetting.ApiKey,
		Prompt:   newSetting.Prompt,
		BaseURL:  httpUtil.TrimURL(newSetting.BaseURL),
	}

	return settingService.transactor.Run(ctx, func(ctx context.Context) error {
		// 序列化为 JSON
		settingToJSON, err := jsonUtil.JSONMarshal(setting)
		if err != nil {
			return err
		}
		settingToJSONString := string(settingToJSON)

		if err := settingService.keyvalueRepository.UpdateKeyValue(ctx, commonModel.AgentSettingKey, settingToJSONString); err != nil {
			return err
		}

		return nil
	})
}
