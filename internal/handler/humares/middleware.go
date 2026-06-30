// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package humares

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-gonic/gin"
	i18nUtil "github.com/lin-snow/ech0/internal/i18n"
	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
)

type ctxKey int

const localizerKey ctxKey = iota

// injectLocalizer 是注册到 huma.API 的全局中间件：让 handler 能在 context 上取到本请求的
// localizer，供 OK/Err 本地化使用（huma handler 拿不到 *gin.Context）。
//
// 关键一：必须写进 gctx.Request 的 context（与 RequireAuth 注入 viewer、StashMeta 注入元数据
// 同一条链），而不能用 huma.WithValue 包裹 ctx——后者会在 auth 之前**快照** ctx.Context()
// 并覆盖之，导致之后 attach 到 gctx.Request 的 viewer/meta 对 handler 不可见。
//
// 关键二：注入的是**惰性 provider**而非 localizer 实例。全局中间件先于 per-operation 的
// 鉴权中间件运行，而 RequireAuth/OptionalAuth 会通过 ApplyUserLocaleFromUserID 在 gin Keys
// 里换上用户偏好 locale。若此处快照实例，handler 拿到的会是鉴权前的旧 localizer；改为在
// 使用时（localizerFrom）才 LocalizerFromGin(gctx) 解析，即可拿到鉴权后的最新值，并与
// installErrorModel 的实时解析保持一致。gctx 整个请求内是同一指针，提前捕获是安全的。
func injectLocalizer(ctx huma.Context, next func(huma.Context)) {
	gctx := humagin.Unwrap(ctx)
	provider := func() *goi18n.Localizer { return i18nUtil.LocalizerFromGin(gctx) }
	gctx.Request = gctx.Request.WithContext(context.WithValue(gctx.Request.Context(), localizerKey, provider))
	next(ctx)
}

// localizerFrom 从 context 取回 injectLocalizer 注入的 provider 并在**使用时**解析 localizer；
// 缺失时返回 nil，i18nUtil.Localize 对 nil localizer 会回退到默认文案。
func localizerFrom(ctx context.Context) *goi18n.Localizer {
	if provider, ok := ctx.Value(localizerKey).(func() *goi18n.Localizer); ok {
		return provider()
	}
	return nil
}

// Bridge 把一个 gin 中间件适配成 Huma operation 中间件，用于**原样复用**现有的
// RequireAuth / RequireScopes / RequireAudience / 限流 等鉴权中间件，零分叉。
//
// 安全性依据（已核实 humagin 源码）：humagin 把 huma handler 注册为路由组上**唯一**的
// gin handler，op 中间件在 huma dispatch 内部运行。因此被桥接中间件内部的 ctx.Next()
// 是无害空操作（其后已无 gin handler）；若中间件 Abort（已写出本地化拒绝响应），
// 则不再进入 huma 下游。RequireAuth 注入的 viewer 顺着 gctx.Request.Context() 自然流入 handler。
//
// 约束（重要）：仅支持「放行或 Abort」型中间件（auth / scope / audience / 限流）。Bridge 先把
// h(gctx) 跑到底、再调 next(ctx) 进入 handler，所以**严禁桥接在 ctx.Next() 之后还做事的中间件**
// （计时 / metrics / 响应改写等），其 post-Next 尾部会先于 handler 执行，静默得到错误结果且无
// 编译或测试信号。这类横切需求请走 Huma 原生中间件，不要经 Bridge。
func Bridge(h gin.HandlerFunc) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		gctx := humagin.Unwrap(ctx)
		h(gctx)
		if gctx.IsAborted() {
			return
		}
		next(ctx)
	}
}
