// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/lin-snow/ech0/internal/handler/humares"
	"github.com/lin-snow/ech0/internal/middleware"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	authService "github.com/lin-snow/ech0/internal/service/auth"
)

// posture 是一个 JSON 端点的「认证(Authn) + 授权(Authz)」姿态。它一处同时产出：
//   - OpenAPI 的 Security 声明（给文档）
//   - 运行时中间件链（给拦截，复用现有 gin 中间件经 humares.Bridge 适配）
//
// 二者由同一个构造函数生成，scope 不可能再写得对不上。authn 维度 = public/optional/secured，
// authz 维度 = secured 的 scopes 参数与 .audience()。
type posture struct {
	security    []map[string][]string
	middlewares huma.Middlewares
}

// public：不认证、不授权——任何人可访问。
func public() posture { return posture{} }

// optional：可选认证(Authn)——无 token 按匿名继续，携带有效 token 则识别用户身份；不做授权。
func optional(revoker authService.TokenRevoker) posture {
	return posture{middlewares: huma.Middlewares{
		humares.Bridge(middleware.NoCache()),
		humares.Bridge(middleware.OptionalAuth(revoker)),
	}}
}

// secured：必须认证(Authn)；附带 scopes 时再要求 scope 授权(Authz)，不带即「仅认证」。
func secured(revoker authService.TokenRevoker, scopes ...string) posture {
	mws := huma.Middlewares{
		humares.Bridge(middleware.NoCache()),
		humares.Bridge(middleware.RequireAuth(revoker)),
	}
	if len(scopes) > 0 {
		mws = append(mws, humares.Bridge(middleware.RequireScopes(scopes...)))
	}
	return posture{security: humares.Secured(scopes...), middlewares: mws}
}

// audience：在 secured 之上叠加 audience 授权(Authz)，用于集成令牌端点。
func (p posture) audience(auds ...string) posture {
	p.middlewares = append(p.middlewares, humares.Bridge(middleware.RequireAudience(auds...)))
	return p
}

// noCache 返回「仅 NoCache」的中间件，给公开但敏感的端点（如注册）在 op 字面量里显式声明。
func noCache() huma.Middlewares {
	return huma.Middlewares{humares.Bridge(middleware.NoCache())}
}

// register 注册一个 JSON 端点：套用 posture（Security + 鉴权中间件），并把中立 handler 经
// humares.Wrap 折成统一信封后交给 huma.Register。
//
// 中间件顺序：posture 的（认证/授权）在前，op 字面量自带的（如限速、评论的 StashMeta 等
// 非鉴权中间件）在后——与重构前一致。
func register[I, T any](api huma.API, p posture, op huma.Operation, h func(context.Context, *I) (commonModel.Result[T], error)) {
	op.Security = p.security
	if len(p.middlewares) > 0 || len(op.Middlewares) > 0 {
		mws := make(huma.Middlewares, 0, len(p.middlewares)+len(op.Middlewares))
		mws = append(mws, p.middlewares...)
		mws = append(mws, op.Middlewares...)
		op.Middlewares = mws
	}
	huma.Register(api, op, humares.Wrap(h))
}
