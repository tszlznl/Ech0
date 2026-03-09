package model

// Status 表示系统初始化状态。
type Status struct {
	Initialized bool `json:"initialized"`
	OwnerExists bool `json:"owner_exists"`
}
