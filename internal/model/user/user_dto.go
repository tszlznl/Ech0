package model

// UserInfoDto 用户信息数据传输对象
//
// swagger:model UserInfoDto
type UserInfoDto struct {
	// 用户名
	// example: linsnow
	Username string `json:"username"`

	// 密码
	// example: 123456
	Password string `json:"password"`

	// 邮箱
	// example: owner@example.com
	Email string `json:"email"`

	// 是否为管理员
	// example: false
	IsAdmin bool `json:"is_admin"`

	// 是否为Owner
	// example: false
	IsOwner bool `json:"is_owner"`

	// 头像地址
	// example: https://example.com/avatar.png
	Avatar string `json:"avatar"`

	// 头像文件ID（用于确认临时文件转正）
	// example: 0195e2a7-54a9-7bcf-8df5-6d81d671f5c7
	AvatarFileID string `json:"avatar_file_id"`

	// 语言偏好
	// example: zh-CN
	Locale string `json:"locale"`
}

// OAuthInfoDto OAuth2 信息数据传输对象
type OAuthInfoDto struct {
	Provider string `json:"provider"`
	UserID   string `json:"user_id"`
	OAuthID  string `json:"oauth_id"`
	Issuer   string `json:"issuer"`
	AuthType string `json:"auth_type"`
}
