package model

// LoginDto 是用户登录时的请求数据传输对象
type LoginDto struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterDto 是用户注册时的请求数据传输对象
type RegisterDto struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email"`
	// Locale 仅在初始化 Owner 时使用：作为部署者偏好的语言写入用户记录与站点默认语言。
	// 普通注册流程会忽略此字段。
	Locale string `json:"locale,omitempty"`
}
