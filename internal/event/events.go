// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package event 是事件系统的纯词汇表：事件结构体及其自描述方法（EventName / OrderingKey）。
// 它只依赖领域 model 与标准库，绝不依赖 busen / 日志 / 任何 service —— 发布方与订阅方都只
// 指向这里。总线机制（pkg/busen 接线、Emit、webhook 桥接、生命周期）独立在 internal/event/bus。
package event

import (
	commentModel "github.com/lin-snow/ech0/internal/model/comment"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	userModel "github.com/lin-snow/ech0/internal/model/user"
)

// Named 是事件的稳定外部名（用作 webhook 观察的 topic / X-Ech0-Event）。所有事件都实现它。
type Named interface{ EventName() string }

// Keyed 是可选的局部有序键：busen 对 async 订阅者按 per-key 保序。实现它的事件在发布时会带上
// WithKey；未实现则不保序。
type Keyed interface{ OrderingKey() string }

type (
	UserCreated struct{ User userModel.User }
	UserUpdated struct{ User userModel.User }
	UserDeleted struct{ User userModel.User }

	EchoCreated struct {
		Echo echoModel.Echo
		User userModel.User
	}
	EchoUpdated struct {
		Echo echoModel.Echo
		User userModel.User
	}
	EchoDeleted struct {
		Echo echoModel.Echo
		User userModel.User
	}

	CommentCreated       struct{ Comment commentModel.Comment }
	CommentStatusUpdated struct{ Comment commentModel.Comment }
	CommentDeleted       struct{ Comment commentModel.Comment }

	ResourceUploaded struct {
		User     userModel.User
		FileName string
		URL      string
		Size     int64
		Type     string
		// Key 是存储 key，仅用于发布时的局部有序（见 OrderingKey）；json:"-" 保证 webhook 载荷不变。
		Key string `json:"-"`
	}

	SystemSnapshot struct {
		Info string
		Size int64
	}
	SystemExport struct {
		Info string
		Size int64
	}

	UpdateSnapshotSchedule struct {
		Schedule settingModel.SnapshotSchedule
	}
)

// EventName —— 稳定外部名，必须与历史 topic 字符串逐字一致（webhook 的 topic 字段兼容）。
func (UserCreated) EventName() string            { return "user.created" }
func (UserUpdated) EventName() string            { return "user.updated" }
func (UserDeleted) EventName() string            { return "user.deleted" }
func (EchoCreated) EventName() string            { return "echo.created" }
func (EchoUpdated) EventName() string            { return "echo.updated" }
func (EchoDeleted) EventName() string            { return "echo.deleted" }
func (CommentCreated) EventName() string         { return "comment.created" }
func (CommentStatusUpdated) EventName() string   { return "comment.status.updated" }
func (CommentDeleted) EventName() string         { return "comment.deleted" }
func (ResourceUploaded) EventName() string       { return "resource.uploaded" }
func (SystemSnapshot) EventName() string         { return "system.snapshot" }
func (SystemExport) EventName() string           { return "system.export" }
func (UpdateSnapshotSchedule) EventName() string { return "system.snapshot_schedule.updated" }

// OrderingKey —— 局部有序键，仅实现于历史上带 WithKey 的事件。
func (e UserCreated) OrderingKey() string          { return e.User.ID }
func (e UserUpdated) OrderingKey() string          { return e.User.ID }
func (e UserDeleted) OrderingKey() string          { return e.User.ID }
func (e EchoCreated) OrderingKey() string          { return e.Echo.ID }
func (e EchoUpdated) OrderingKey() string          { return e.Echo.ID }
func (e EchoDeleted) OrderingKey() string          { return e.Echo.ID }
func (e CommentCreated) OrderingKey() string       { return e.Comment.ID }
func (e CommentStatusUpdated) OrderingKey() string { return e.Comment.ID }
func (e CommentDeleted) OrderingKey() string       { return e.Comment.ID }
func (e ResourceUploaded) OrderingKey() string     { return e.Key }
