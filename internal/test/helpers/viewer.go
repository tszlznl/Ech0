// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package helpers

import (
	"context"

	"github.com/lin-snow/ech0/pkg/viewer"
)

// CtxAsUser 返回带普通用户身份的 context（service 层用 viewer.MustFromContext 取身份）。
//
// 说明：管理员/owner 身份不在 viewer 上，而是用户 DB 记录的 IsAdmin/IsOwner 字段决定。
// 单测里应通过 fixture 建管理员用户 + mock 的 commonService 返回该用户来表达「管理员」。
func CtxAsUser(userID string) context.Context {
	return viewer.WithContext(context.Background(), viewer.NewUserViewer(userID))
}

// CtxAsToken 返回带访问令牌身份（scope/audience/jti）的 context，用于 scope/audience 相关测试。
func CtxAsToken(userID, tokenType string, scopes, audience []string, jti string) context.Context {
	return viewer.WithContext(
		context.Background(),
		viewer.NewUserViewerWithToken(userID, tokenType, scopes, audience, jti),
	)
}

// CtxAnonymous 返回匿名（Noop）身份 context。
func CtxAnonymous() context.Context {
	return viewer.WithContext(context.Background(), viewer.NewNoopViewer())
}
