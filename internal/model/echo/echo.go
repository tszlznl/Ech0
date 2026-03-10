package model

import (
	"time"

	fileModel "github.com/lin-snow/ech0/internal/model/file"
	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
	"gorm.io/gorm"
)

// Echo 定义Echo实体
type Echo struct {
	ID            string               `gorm:"type:char(36);primaryKey"                      json:"id"`
	Content       string               `gorm:"type:text;not null"                            json:"content"`
	Username      string               `gorm:"type:varchar(100)"                             json:"username,omitempty"`
	EchoFiles     []fileModel.EchoFile `gorm:"foreignKey:EchoID;constraint:OnDelete:CASCADE" json:"echo_files,omitempty"`
	Layout        string               `gorm:"type:varchar(50);default:'waterfall'"          json:"layout,omitempty"`
	Private       bool                 `gorm:"default:false;index:idx_echos_private_created,priority:1" json:"private"`
	UserID        string               `gorm:"type:char(36);not null;index"                  json:"user_id"`
	Extension     *EchoExtension       `gorm:"foreignKey:EchoID;constraint:OnDelete:CASCADE" json:"extension,omitempty"`
	Tags          []Tag                `gorm:"many2many:echo_tags;"                          json:"tags,omitempty"`
	FavCount      int                  `gorm:"default:0"                                     json:"fav_count"`
	CreatedAt     time.Time            `gorm:"index:idx_echos_private_created,priority:2"    json:"created_at"`
}

type EchoExtension struct {
	ID        string                 `gorm:"type:char(36);primaryKey"      json:"id"`
	EchoID    string                 `gorm:"type:char(36);not null;uniqueIndex" json:"echo_id"`
	Type      string                 `gorm:"type:varchar(100);not null"    json:"type"`
	Payload   map[string]interface{} `gorm:"serializer:json;type:text;not null" json:"payload"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// Tag 定义Tag实体
type Tag struct {
	ID         string    `gorm:"type:char(36);primaryKey"              json:"id"`
	Name       string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"name"`
	UsageCount int       `gorm:"default:0"                             json:"usage_count"`
	CreatedAt  time.Time `                                             json:"created_at"`
}

// EchoTag 纯关系表，联合主键
type EchoTag struct {
	EchoID string `gorm:"type:char(36);primaryKey"`
	TagID  string `gorm:"type:char(36);primaryKey;index"`
}

func (e *Echo) BeforeCreate(_ *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuidUtil.MustNewV7()
	}
	return nil
}

func (t *Tag) BeforeCreate(_ *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuidUtil.MustNewV7()
	}
	return nil
}

func (e *EchoExtension) BeforeCreate(_ *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuidUtil.MustNewV7()
	}
	return nil
}

const (
	Extension_MUSIC      = "MUSIC"
	Extension_VIDEO      = "VIDEO"
	Extension_GITHUBPROJ = "GITHUBPROJ"
	Extension_WEBSITE    = "WEBSITE"

	LayoutWaterfall  = "waterfall"
	LayoutGrid       = "grid"
	LayoutHorizontal = "horizontal"
	LayoutCarousel   = "carousel"
)
