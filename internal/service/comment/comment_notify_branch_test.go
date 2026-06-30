// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service_test

import (
	"context"
	"testing"

	commentModel "github.com/lin-snow/ech0/internal/model/comment"
	"github.com/lin-snow/ech0/internal/test/helpers"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// 这些用例只覆盖通知函数里同步的「提前返回」分支：邮件开关开着、shouldNotify=true，
// 但收件人无效 / owner 邮箱缺失 / 被回复者即 owner —— 都在 spawn 异步发信 goroutine 之前返回，
// 因此完全确定、无需 sleep、不会触发 mailer。

// mailEnabledSetting 在基线设置上打开邮件通知总开关。
func mailEnabledSetting() commentModel.SystemSetting {
	s := enabledSetting()
	s.EmailNotify.Enabled = true
	return s
}

// notifyOwnerAsync：status 类通知用「评论者邮箱」作收件人；邮箱无效时跳过发信。
func TestUpdateCommentStatus_NotifySkipsInvalidRecipient(t *testing.T) {
	d := newDeps(t)
	expectAdmin(t, d, "admin-1")
	d.repo.EXPECT().
		UpdateCommentStatus(mock.Anything, "c-1", commentModel.StatusRejected).
		Return(nil).
		Once()
	// 回读到的评论邮箱为空 => parseValidEmail 失败 => notifyOwnerAsync 在 spawn 前返回。
	d.repo.EXPECT().
		GetCommentByID(mock.Anything, "c-1").
		Return(commentModel.Comment{ID: "c-1", Status: commentModel.StatusRejected, Email: ""}, nil).
		Once()
	d.expectSetting(t, mailEnabledSetting())

	err := d.service().UpdateCommentStatus(helpers.CtxAsUser("admin-1"), "c-1", commentModel.StatusRejected)
	require.NoError(t, err)
}

// notifyOwnerAsync：created 类通知发给 owner；owner 邮箱缺失时在 spawn 前返回。
func TestCreateComment_NotifyOwnerSkipsWhenOwnerEmailMissing(t *testing.T) {
	helpers.SetJWTSecret(t, testSecret)
	d := newDeps(t)
	s := mailEnabledSetting()
	s.RequireApproval = false
	d.expectSetting(t, s)
	d.repo.EXPECT().
		CountByIPWithin(mock.Anything, mock.Anything, mock.Anything).
		Return(int64(0), nil)
	d.repo.EXPECT().
		CountByEmailWithin(mock.Anything, mock.Anything, mock.Anything).
		Return(int64(0), nil)
	d.repo.EXPECT().
		ExistsRecentDuplicate(
			mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything,
		).
		Return(false, nil).
		Once()
	d.repo.EXPECT().
		CreateComment(mock.Anything, mock.Anything).
		Run(func(_ context.Context, c *commentModel.Comment) { c.ID = "g-1" }).
		Return(nil).
		Once()
	// resolveOwnerEmail：owner 邮箱为空 => 返回错误 => notifyOwnerAsync 在 spawn 前返回。
	d.common.EXPECT().GetOwner().Return(helpers.NewUser(), nil).Once()

	res, err := d.service().CreateComment(helpers.CtxAnonymous(), testIP, "ua",
		&commentModel.CreateCommentDto{
			EchoID:    "echo-1",
			Content:   "hello",
			Nickname:  "Guest",
			Email:     "guest@example.com",
			FormToken: freshToken(),
		})
	require.NoError(t, err)
	assert.Equal(t, "g-1", res.ID)
}

// notifyReplyTargetAsync：被回复评论邮箱无效时跳过（owner 自评不触发「新评论」通知，故只走回复通知）。
func TestCreateComment_ReplyNotifySkipsInvalidParentEmail(t *testing.T) {
	helpers.SetJWTSecret(t, testSecret)
	d := newDeps(t)
	d.expectSetting(t, mailEnabledSetting())

	owner := helpers.NewUser(helpers.AsOwner)
	owner.ID = "owner-1"
	d.common.EXPECT().
		CommonGetUserByUserId(mock.Anything, "owner-1").
		Return(owner, nil).
		Once()
	// 父评论被读两次：resolveParentID 校验 + notifyReplyTargetAsync 取收件邮箱。
	d.repo.EXPECT().
		GetCommentByID(mock.Anything, "parent-1").
		Return(commentModel.Comment{
			ID:     "parent-1",
			EchoID: "echo-1",
			Status: commentModel.StatusApproved,
			Email:  "", // 无效 => parseValidEmail 失败 => 在 spawn 前返回
		}, nil).
		Times(2)
	d.repo.EXPECT().
		ExistsRecentDuplicate(
			mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything,
		).
		Return(false, nil).
		Once()
	d.repo.EXPECT().
		CreateComment(mock.Anything, mock.Anything).
		Run(func(_ context.Context, c *commentModel.Comment) { c.ID = "child-1" }).
		Return(nil).
		Once()

	res, err := d.service().CreateComment(helpers.CtxAsUser("owner-1"), testIP, "ua",
		&commentModel.CreateCommentDto{
			EchoID:    "echo-1",
			Content:   "a reply",
			ParentID:  "parent-1",
			FormToken: freshToken(),
		})
	require.NoError(t, err)
	assert.Equal(t, "child-1", res.ID)
}

// notifyReplyTargetAsync：被回复者恰好是 owner 时跳过（已由「新评论」通知覆盖）。
func TestCreateComment_ReplyNotifySkipsWhenTargetIsOwner(t *testing.T) {
	helpers.SetJWTSecret(t, testSecret)
	d := newDeps(t)
	d.expectSetting(t, mailEnabledSetting())

	owner := helpers.NewUser(helpers.AsOwner)
	owner.ID = "owner-1"
	d.common.EXPECT().
		CommonGetUserByUserId(mock.Anything, "owner-1").
		Return(owner, nil).
		Once()
	d.repo.EXPECT().
		GetCommentByID(mock.Anything, "parent-1").
		Return(commentModel.Comment{
			ID:     "parent-1",
			EchoID: "echo-1",
			Status: commentModel.StatusApproved,
			Email:  "shared@example.com",
		}, nil).
		Times(2)
	d.repo.EXPECT().
		ExistsRecentDuplicate(
			mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything,
		).
		Return(false, nil).
		Once()
	d.repo.EXPECT().
		CreateComment(mock.Anything, mock.Anything).
		Run(func(_ context.Context, c *commentModel.Comment) { c.ID = "child-2" }).
		Return(nil).
		Once()
	// 被回复邮箱 == owner 邮箱 => 在 spawn 前返回。
	ownerWithEmail := helpers.NewUser(helpers.AsOwner)
	ownerWithEmail.Email = "shared@example.com"
	d.common.EXPECT().GetOwner().Return(ownerWithEmail, nil).Once()

	res, err := d.service().CreateComment(helpers.CtxAsUser("owner-1"), testIP, "ua",
		&commentModel.CreateCommentDto{
			EchoID:    "echo-1",
			Content:   "a reply",
			ParentID:  "parent-1",
			FormToken: freshToken(),
		})
	require.NoError(t, err)
	assert.Equal(t, "child-2", res.ID)
}
