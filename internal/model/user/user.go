package model

import (
	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
	"gorm.io/gorm"
)

const (
	USER_NOT_EXISTS_ID = ""
)

// User 定义用户实体
type User struct {
	ID       string `gorm:"type:char(36);primaryKey" json:"id"`
	Username string `gorm:"size:255;not null;unique" json:"username"`
	Password string `gorm:"size:255;not null"        json:"-"`
	IsAdmin  bool   `gorm:"bool"                     json:"is_admin"`
	IsOwner  bool   `gorm:"bool"                     json:"is_owner"`
	Avatar   string `gorm:"size:255"                 json:"avatar"`
}

type OAuthBinding struct {
	ID       string `gorm:"type:char(36);primaryKey" json:"id"`
	UserID   string `gorm:"type:char(36);not null;index" json:"user_id"` // Ech0 用户 ID
	Provider string `gorm:"size:64;not null;index"  json:"provider"`     // 例如 "github"，"google"，"qq"，"custom"，"oidc"
	OAuthID  string `gorm:"size:255;not null;index" json:"oauth_id"`     // OAuth2: oauth_id, OIDC: sub, 第三方平台的用户ID
	Issuer   string `gorm:"size:255;"               json:"issuer"`       // OIDC: issuer
	AuthType string `gorm:"size:64;"                json:"auth_type"`    // OAuth2: null || 'oauth2', OIDC: not null && 'oidc'
}

func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuidUtil.MustNewV7()
	}
	return nil
}

func (o *OAuthBinding) BeforeCreate(_ *gorm.DB) error {
	if o.ID == "" {
		o.ID = uuidUtil.MustNewV7()
	}
	return nil
}
