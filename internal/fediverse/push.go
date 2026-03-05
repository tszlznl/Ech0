package fediverse

import (
	"encoding/json"
	"errors"
	"fmt"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	httpUtil "github.com/lin-snow/ech0/internal/util/http"
	jsonUtil "github.com/lin-snow/ech0/internal/util/json"
)

// PushEchoToFediverse 将 Echo 推送到联邦网络
func (core *FediverseCore) PushEchoToFediverse(userId uint, echo echoModel.Echo) error {
	// 检查是否开启了联邦网络功能
	var fediverseSetting settingModel.FediverseSetting
	if fediverseSettingJSON, err := core.keyvalueRepo.GetKeyValue(commonModel.FediverseSettingKey); err == nil {
		if err := jsonUtil.JSONUnmarshal([]byte(fediverseSettingJSON), &fediverseSetting); err != nil {
			return err
		}
	} else {
		return err
	}

	if !fediverseSetting.Enable {
		return nil
	}

	if echo.Private {
		return nil
	}

	// 获取用户
	user, err := core.userRepository.GetUserByID(int(userId))
	if err != nil {
		return err
	}

	// 获取粉丝列表
	followers, err := core.repo.GetFollowers(user.ID)
	if err != nil {
		return err
	}
	if len(followers) == 0 {
		return nil
	}

	// 获取 Actor 和 setting
	actor, setting, err := core.BuildActor(&user)
	if err != nil {
		return err
	}

	serverURL, err := NormalizeServerURL(setting.ServerURL)
	if err != nil {
		return err
	}

	activity := core.ConvertEchoToActivity(&echo, &actor, serverURL)
	object := core.ConvertEchoToObject(&echo, &actor, serverURL)

	activityMap := map[string]any{}
	activityBytes, err := json.Marshal(activity)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(activityBytes, &activityMap); err != nil {
		return err
	}

	objectMap := map[string]any{}
	objectBytes, err := json.Marshal(object)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(objectBytes, &objectMap); err != nil {
		return err
	}

	activityMap["object"] = objectMap

	payloadBytes, err := json.Marshal(activityMap)
	if err != nil {
		return err
	}

	var errs []error
	// 推送到每个粉丝的Inbox
	for _, follower := range followers {
		inboxURL, err := core.FetchRemoteActorInbox(follower.ActorID)
		if err != nil {
			errs = append(errs, fmt.Errorf("fetch inbox for %s: %w", follower.ActorID, err))
			continue
		}

		if err := httpUtil.PostActivity(payloadBytes, inboxURL, actor.ID); err != nil {
			errs = append(errs, fmt.Errorf("post activity to %s: %w", inboxURL, err))
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
