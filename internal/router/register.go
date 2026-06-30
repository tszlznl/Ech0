// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/lin-snow/ech0/internal/handler/humares"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
)

// reg 注册一个 Huma operation：把中立 handler（返回 commonModel.Result[T]，自带成功提示）
// 经 humares.Wrap 折成统一信封后交给 huma.Register。无数据端点即 T=any。
func reg[I, T any](api huma.API, op huma.Operation, h func(context.Context, *I) (commonModel.Result[T], error)) {
	huma.Register(api, op, humares.Wrap(h))
}
