package service

import (
	"context"
	"encoding/json"
	"errors"

	contracts "github.com/lin-snow/ech0/internal/event/contracts"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
	fmtUtil "github.com/lin-snow/ech0/internal/util/format"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"github.com/lin-snow/ech0/pkg/viewer"
	"go.uber.org/zap"
)

// GetBackupScheduleSetting 获取备份计划
func (settingService *SettingService) GetBackupScheduleSetting(
	setting *model.BackupSchedule,
) error {
	// 鉴权
	// user, err := settingService.commonService.CommonGetUserByUserId(userid)
	// if err != nil {
	// 	return err
	// }
	// if !user.IsAdmin {
	// 	return errors.New(commonModel.NO_PERMISSION_DENIED)
	// }

	return settingService.transactor.Run(context.Background(), func(ctx context.Context) error {
		backupSchedule, err := settingService.keyvalueRepository.GetKeyValue(
			ctx,
			commonModel.BackupScheduleKey,
		)
		if err != nil {
			// 数据库中不存在数据，手动添加初始数据
			setting.Enable = false
			// 默认每周日凌晨2点备份
			setting.CronExpression = "0 2 * * 0"

			// 序列化为 JSON
			settingToJSON, err := json.Marshal(setting)
			if err != nil {
				return err
			}
			if err := settingService.keyvalueRepository.AddKeyValue(ctx, commonModel.BackupScheduleKey, string(settingToJSON)); err != nil {
				return err
			}

			return nil
		}

		if err := json.Unmarshal([]byte(backupSchedule), setting); err != nil {
			return err
		}

		return nil
	})
}

// UpdateBackupScheduleSetting 更新备份计划
func (settingService *SettingService) UpdateBackupScheduleSetting(
	ctx context.Context,
	newSetting *model.BackupScheduleDto,
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

	var updated model.BackupSchedule
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
		if err := settingService.keyvalueRepository.UpdateKeyValue(ctx, commonModel.BackupScheduleKey, settingToJSONString); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	// 在事务提交后再发布事件，避免回滚时出现幽灵事件。
	if err := settingService.publisher.BackupScheduleUpdated(
		context.Background(),
		contracts.UpdateBackupScheduleEvent{Schedule: updated},
	); err != nil {
		logUtil.GetLogger().
			Error("Failed to publish update backup schedule event", zap.Error(err))
	}
	return nil
}
