// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package event

import (
	"testing"

	commentModel "github.com/lin-snow/ech0/internal/model/comment"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEventName 锁定每个事件的稳定外部名（webhook 的 topic / X-Ech0-Event）。
// 这些字符串是 webhook 兼容契约，任何改动都属破坏性变更，必须显式失败。
func TestEventName(t *testing.T) {
	cases := []struct {
		name string
		ev   Named
		want string
	}{
		{"UserCreated", UserCreated{}, "user.created"},
		{"UserUpdated", UserUpdated{}, "user.updated"},
		{"UserDeleted", UserDeleted{}, "user.deleted"},
		{"EchoCreated", EchoCreated{}, "echo.created"},
		{"EchoUpdated", EchoUpdated{}, "echo.updated"},
		{"EchoDeleted", EchoDeleted{}, "echo.deleted"},
		{"CommentCreated", CommentCreated{}, "comment.created"},
		{"CommentStatusUpdated", CommentStatusUpdated{}, "comment.status.updated"},
		{"CommentDeleted", CommentDeleted{}, "comment.deleted"},
		{"ResourceUploaded", ResourceUploaded{}, "resource.uploaded"},
		{"SystemSnapshot", SystemSnapshot{}, "system.snapshot"},
		{"SystemExport", SystemExport{}, "system.export"},
		{"UpdateSnapshotSchedule", UpdateSnapshotSchedule{}, "system.snapshot_schedule.updated"},
	}

	// 守卫事件总数：新增/删除事件时此处必须同步更新，避免遗漏 topic 契约锁定。
	require.Len(t, cases, 13, "expected exactly 13 named events")

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.ev.EventName())
		})
	}
}

// TestEventNamesUnique 保证所有 topic 字符串互不重复——webhook 路由按名分发，重名会串台。
func TestEventNamesUnique(t *testing.T) {
	names := []Named{
		UserCreated{},
		UserUpdated{},
		UserDeleted{},
		EchoCreated{},
		EchoUpdated{},
		EchoDeleted{},
		CommentCreated{},
		CommentStatusUpdated{},
		CommentDeleted{},
		ResourceUploaded{},
		SystemSnapshot{},
		SystemExport{},
		UpdateSnapshotSchedule{},
	}

	seen := make(map[string]string, len(names))
	for _, n := range names {
		topic := n.EventName()
		assert.NotEmpty(t, topic, "topic must not be empty")
		if prev, dup := seen[topic]; dup {
			t.Errorf("duplicate topic %q used by multiple events (also %q)", topic, prev)
			continue
		}
		seen[topic] = topic
	}
	assert.Len(t, seen, len(names), "every event must have a distinct topic")
}

// TestOrderingKey 锁定局部有序键：busen 对 async 订阅者按 per-key 保序。
// 仅历史上带 WithKey 的 10 个事件实现 Keyed，键值取对应资源 ID。
func TestOrderingKey(t *testing.T) {
	const (
		userID    = "user-key-0001"
		echoID    = "echo-key-0001"
		commentID = "comment-key-0001"
		storeKey  = "uploads/2026/abc.png"
	)

	cases := []struct {
		name string
		ev   Keyed
		want string
	}{
		{"UserCreated", UserCreated{User: userModel.User{ID: userID}}, userID},
		{"UserUpdated", UserUpdated{User: userModel.User{ID: userID}}, userID},
		{"UserDeleted", UserDeleted{User: userModel.User{ID: userID}}, userID},
		{"EchoCreated", EchoCreated{Echo: echoModel.Echo{ID: echoID}}, echoID},
		{"EchoUpdated", EchoUpdated{Echo: echoModel.Echo{ID: echoID}}, echoID},
		{"EchoDeleted", EchoDeleted{Echo: echoModel.Echo{ID: echoID}}, echoID},
		{"CommentCreated", CommentCreated{Comment: commentModel.Comment{ID: commentID}}, commentID},
		{"CommentStatusUpdated", CommentStatusUpdated{Comment: commentModel.Comment{ID: commentID}}, commentID},
		{"CommentDeleted", CommentDeleted{Comment: commentModel.Comment{ID: commentID}}, commentID},
		{"ResourceUploaded", ResourceUploaded{Key: storeKey}, storeKey},
	}

	// 守卫 Keyed 事件总数。
	require.Len(t, cases, 10, "expected exactly 10 keyed events")

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.ev.OrderingKey())
		})
	}
}

// TestOrderingKey_EmptyWhenIDUnset 验证键直接透传资源 ID：未设置 ID 时键为空（不保序）。
func TestOrderingKey_EmptyWhenIDUnset(t *testing.T) {
	assert.Empty(t, EchoCreated{}.OrderingKey())
	assert.Empty(t, UserCreated{}.OrderingKey())
	assert.Empty(t, CommentCreated{}.OrderingKey())
	// ResourceUploaded 的有序键取存储 Key 而非 FileName/URL。
	assert.Empty(t, ResourceUploaded{FileName: "a.png", URL: "http://x/a.png"}.OrderingKey())
}

// TestSystemEventsNotKeyed 回归：系统类事件不应实现 Keyed（历史上不带 WithKey，不保序）。
func TestSystemEventsNotKeyed(t *testing.T) {
	cases := []struct {
		name string
		ev   Named
	}{
		{"SystemSnapshot", SystemSnapshot{}},
		{"SystemExport", SystemExport{}},
		{"UpdateSnapshotSchedule", UpdateSnapshotSchedule{Schedule: settingModel.SnapshotSchedule{}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, ok := tc.ev.(Keyed)
			assert.False(t, ok, "%s must NOT implement Keyed", tc.name)
		})
	}
}

// TestKeyedEventsAlsoNamed 编译期+运行期双保险：每个 Keyed 事件同时是 Named。
func TestKeyedEventsAlsoNamed(t *testing.T) {
	keyed := []Keyed{
		UserCreated{},
		UserUpdated{},
		UserDeleted{},
		EchoCreated{},
		EchoUpdated{},
		EchoDeleted{},
		CommentCreated{},
		CommentStatusUpdated{},
		CommentDeleted{},
		ResourceUploaded{},
	}
	for _, k := range keyed {
		n, ok := k.(Named)
		require.True(t, ok, "keyed event must also be Named")
		assert.NotEmpty(t, n.EventName())
	}
}
