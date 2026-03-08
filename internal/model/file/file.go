package model

import (
	"time"

	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
	"gorm.io/gorm"
)

type File struct {
	ID string `gorm:"type:char(36);primaryKey" json:"id"`

	// 存储键（本地文件名或对象存储 object key）
	Key string `gorm:"type:varchar(500);not null;uniqueIndex:idx_file_route,priority:4" json:"key"`

	StorageType string `gorm:"type:varchar(20);not null;uniqueIndex:idx_file_route,priority:1" json:"storage_type"` // local|object
	Provider    string `gorm:"type:varchar(50);uniqueIndex:idx_file_route,priority:2" json:"provider,omitempty"`    // object 存储提供商，如 aws/r2/minio
	Bucket      string `gorm:"type:varchar(120);uniqueIndex:idx_file_route,priority:3" json:"bucket,omitempty"`     // local 可空

	URL         string `gorm:"type:text" json:"url"` // 前端直链快照
	Name        string `gorm:"type:varchar(255)" json:"name"`
	ContentType string `gorm:"type:varchar(100)" json:"content_type,omitempty"`
	Size        int64  `gorm:"default:0" json:"size"`
	Width       int    `gorm:"default:0" json:"width,omitempty"`
	Height      int    `gorm:"default:0" json:"height,omitempty"`

	Category  string    `gorm:"type:varchar(20);index" json:"category"` // image|video|audio|document|file
	UserID    string    `gorm:"type:char(36);index;not null" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

// EchoFile links a File to an Echo with ordering support.
type EchoFile struct {
	ID        string `gorm:"type:char(36);primaryKey"                        json:"id"`
	EchoID    string `gorm:"type:char(36);uniqueIndex:idx_echo_file;not null" json:"echo_id"`
	FileID    string `gorm:"type:char(36);uniqueIndex:idx_echo_file;not null" json:"file_id"`
	File      File   `gorm:"foreignKey:FileID;constraint:OnDelete:CASCADE" json:"file,omitempty"`
	SortOrder int    `gorm:"default:0"                                   json:"sort_order"`
}

func (f *File) BeforeCreate(_ *gorm.DB) error {
	if f.ID == "" {
		f.ID = uuidUtil.MustNewV7()
	}
	return nil
}

func (e *EchoFile) BeforeCreate(_ *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuidUtil.MustNewV7()
	}
	return nil
}
