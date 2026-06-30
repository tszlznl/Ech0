// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package humares

import (
	"context"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
)

// Wrap 把一个**中立** handler（返回 commonModel.Result[T] + error，不认识 Huma）适配成 Huma
// operation 处理函数：成功时本地化 Result 并套进 Envelope[T]，失败时经 Err 映射成统一错误信封。
//
// handler 自己用 commonModel.OK(data[, msg]) 构造 Result（持有成功提示 / message_key），与重构前
// humares.OK 的内部构造逐字节一致；本地化（推导 message_key + 翻译 msg）集中在此。
// 因为返回类型仍是 *Envelope[T]，Huma 反射出的 I/O 泛型类型不变，OpenAPI schema 不变。
//
// 无数据端点即 T=any（handler 返回 commonModel.OK[any](nil, msg)）；需要显式 message_key 的端点
// 由 handler 在 Result 上预设 MessageKey，localizeResult 不会覆盖它。
func Wrap[I, T any](h func(context.Context, *I) (commonModel.Result[T], error)) func(context.Context, *I) (*Envelope[T], error) {
	return func(ctx context.Context, in *I) (*Envelope[T], error) {
		res, err := h(ctx, in)
		if err != nil {
			return nil, Err(ctx, err)
		}
		return &Envelope[T]{Body: localizeResult(ctx, res)}, nil
	}
}
