package model

import (
	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
	"gorm.io/gorm"
)

type Inbox struct {
	ID        string `gorm:"type:char(36);primaryKey;index:idx_inbox_read_created,priority:3" json:"id"` // 收件箱消息ID
	Source    string `gorm:"type:varchar(50);not null" json:"source"`                                    // 消息来源: system/user/agent
	Content   string `gorm:"type:text"                 json:"content"`                                   // 消息内容
	Type      string `gorm:"type:varchar(50);not null" json:"type"`                                      // 消息类型: echo/notification...
	Read      bool   `gorm:"default:false;index:idx_inbox_read_created,priority:1" json:"read"`          // 是否已读
	ReadCount int    `gorm:"default:0"                 json:"read_count"`                                // 已读次数
	Meta      string `gorm:"type:text"                 json:"meta,omitempty"`                            // 额外元数据 (JSON格式)
	ReadAt    int64  `                                 json:"read_at,omitempty"`                         // 已读时间 (Unix时间戳)
	CreatedAt int64  `gorm:"index:idx_inbox_read_created,priority:2" json:"created_at"`                  // 创建时间 (Unix时间戳)
}

func (i *Inbox) BeforeCreate(_ *gorm.DB) error {
	if i.ID == "" {
		i.ID = uuidUtil.MustNewV7()
	}
	return nil
}
