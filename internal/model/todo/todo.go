package model

import (
	"time"

	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
	"gorm.io/gorm"
)

// Todo 定义待办事项实体
type Todo struct {
	ID        string    `gorm:"type:char(36);primaryKey" json:"id"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	UserID    string    `gorm:"type:char(36);not null;index" json:"user_id"`
	Username  string    `gorm:"type:varchar(100)"  json:"username,omitempty"`
	Status    uint      `gorm:"default:0"          json:"status"` // 0:未完成 1:已完成
	CreatedAt time.Time `                          json:"created_at"`
}

func (t *Todo) BeforeCreate(_ *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuidUtil.MustNewV7()
	}
	return nil
}

// Todo 相关状态常量
const (
	Done         = 1 // 待办事项已完成状态
	NotDone      = 0 // 待办事项未完成状态
	MaxTodoCount = 3 // 最大待办事项数量
)
