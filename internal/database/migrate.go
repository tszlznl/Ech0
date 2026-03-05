package database

import (
	"errors"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
)

// migrateImageForeignKeyToEchoID 将旧的 message_id 列迁移为 echo_id 列
func migrateImageForeignKeyToEchoID() error {
	db := GetDB()
	if db == nil {
		return errors.New(commonModel.DATABASE_NOT_INITED)
	}

	migrator := db.Migrator()
	hasOld := migrator.HasColumn(&echoModel.Image{}, "message_id")
	hasNew := migrator.HasColumn(&echoModel.Image{}, "echo_id")
	if hasOld && !hasNew {
		if err := migrator.RenameColumn(&echoModel.Image{}, "message_id", "echo_id"); err != nil {
			return err
		}
	}

	return nil
}

// fixOldEchoLayoutData 为旧数据补充默认的布局值（layout 为 NULL 或空字符串时设为 'waterfall'）
func fixOldEchoLayoutData() error {
	db := GetDB()
	if db == nil {
		return errors.New(commonModel.DATABASE_NOT_INITED)
	}

	// 更新所有 layout 为 NULL 或空字符串的 echo 记录为 'waterfall'
	if err := db.Model(&echoModel.Echo{}).
		Where("layout IS NULL OR layout = ''").
		Update("layout", "waterfall").Error; err != nil {
		return err
	}

	return nil
}

// UpdateMigration 执行旧数据库迁移和数据修复任务
func UpdateMigration() error {
	if err := migrateImageForeignKeyToEchoID(); err != nil {
		return err
	}
	return fixOldEchoLayoutData()
}
