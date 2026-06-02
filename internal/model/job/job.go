// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package model 定义通用 Job（长时有状态作业）的持久化模型与状态/类型常量。
//
// 模型放在 internal/model/job（而非框架包 internal/job），让 AutoMigrate 能像其它
// 领域模型一样以 jobModel 别名引入，不被框架对领域 service 的依赖污染编译图。
package model

// Status 是作业的生命周期状态（无 idle —— 「无作业」由查无行表达）。
type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusSuccess   Status = "success"
	StatusFailed    Status = "failed"
	StatusCancelled Status = "cancelled"
)

// IsTerminal 报告该状态是否为终态（不再流转）。
func (s Status) IsTerminal() bool {
	return s == StatusSuccess || s == StatusFailed || s == StatusCancelled
}

// 作业类型常量：作为 Job 主键 Type 的取值，供 handler/runner 共用。
const (
	TypeReindex   = "reindex"
	TypeMigration = "migration"
)

// Job 是通用作业的持久化行。主键即 Type，结构性保证「每类型单行」：新一次 Submit
// upsert 覆盖旧终态行，无 id、无历史。领域专属的输入/进度/结果序列化进 Payload(JSON)，
// 框架不解析它，只有对应 Runner 与前端认得。
type Job struct {
	Type       string `gorm:"primaryKey;size:64"      json:"type"`
	Status     Status `gorm:"type:varchar(32);index"  json:"status"`
	Phase      string `gorm:"type:varchar(64)"        json:"phase"`
	Error      string `gorm:"type:text"               json:"error"`
	Payload    string `gorm:"type:text"               json:"payload"`
	StartedAt  *int64 `                               json:"started_at"`
	FinishedAt *int64 `                               json:"finished_at"`
	UpdatedAt  int64  `gorm:"autoUpdateTime"          json:"updated_at"`
}

// TableName 固定表名为 jobs。
func (Job) TableName() string {
	return "jobs"
}
