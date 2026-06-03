// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/lin-snow/ech0/internal/event"
	eventbus "github.com/lin-snow/ech0/internal/event/bus"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
	fmtUtil "github.com/lin-snow/ech0/internal/util/format"
	"github.com/lin-snow/ech0/pkg/viewer"
)

// GetSnapshotScheduleSetting 获取定时快照计划
func (settingService *SettingService) GetSnapshotScheduleSetting(
	setting *model.SnapshotSchedule,
) error {
	return settingService.transactor.Run(context.Background(), func(ctx context.Context) error {
		snapshotSchedule, err := settingService.durableKV.Get(
			ctx,
			commonModel.SnapshotScheduleKey,
		)
		if err != nil {
			// 数据库中不存在数据，手动添加初始数据
			setting.Enable = false
			// 默认每周日凌晨2点创建快照
			setting.CronExpression = "0 2 * * 0"

			// 序列化为 JSON
			settingToJSON, err := json.Marshal(setting)
			if err != nil {
				return err
			}
			if err := settingService.durableKV.Set(ctx, commonModel.SnapshotScheduleKey, string(settingToJSON)); err != nil {
				return err
			}

			return nil
		}

		if err := json.Unmarshal([]byte(snapshotSchedule), setting); err != nil {
			return err
		}

		return nil
	})
}

// UpdateSnapshotScheduleSetting 更新定时快照计划
func (settingService *SettingService) UpdateSnapshotScheduleSetting(
	ctx context.Context,
	newSetting *model.SnapshotScheduleDto,
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

	var updated model.SnapshotSchedule
	err = settingService.transactor.Run(ctx, func(ctx context.Context) error {
		updated.Enable = newSetting.Enable
		updated.CronExpression = newSetting.CronExpression

		// 验证 Cron 表达式是否合法
		if err := fmtUtil.ValidateCrontabExpression(updated.CronExpression); err != nil {
			return errors.New(commonModel.INVALID_CRON_EXPRESSION)
		}

		settingToJSON, err := json.Marshal(updated)
		if err != nil {
			return err
		}

		// 将字节切片转换为字符串
		settingToJSONString := string(settingToJSON)
		if err := settingService.durableKV.Set(ctx, commonModel.SnapshotScheduleKey, settingToJSONString); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	// 在事务提交后再发布事件，避免回滚时出现幽灵事件。
	eventbus.Notify(context.Background(), settingService.bus, event.UpdateSnapshotSchedule{Schedule: updated})
	return nil
}
