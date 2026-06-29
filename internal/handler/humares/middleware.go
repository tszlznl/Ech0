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

// injectLocalizer 是注册到 huma.API 的全局中间件：把 gin i18n 中间件已解析好的 localizer
// 注入到 handler 的 context，供 OK/Err 本地化使用（huma handler 拿不到 *gin.Context）。
func injectLocalizer(ctx huma.Context, next func(huma.Context)) {
	loc := i18nUtil.LocalizerFromGin(humagin.Unwrap(ctx))
	next(huma.WithValue(ctx, localizerKey, loc))
}

// localizerFrom 从 context 取回 injectLocalizer 注入的 localizer；缺失时返回 nil，
// i18nUtil.Localize 对 nil localizer 会回退到默认文案。
func localizerFrom(ctx context.Context) *goi18n.Localizer {
	if loc, ok := ctx.Value(localizerKey).(*goi18n.Localizer); ok {
		return loc
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
