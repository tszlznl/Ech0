// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package job

import (
	"context"
	"encoding/json"
	"fmt"
)

// TypedRun 是作者端 typed 的工作函数：直接拿领域 payload 结构体 P。
type TypedRun[P any] func(ctx context.Context, p P, report ReportFunc) (any, error)

// Adapt 把 typed 的 TypedRun[P] 适配成 untyped 的 Runner。异构注册表无法容纳不同
// payload 类型，边界必然擦除——故泛型只放作者端，Unmarshal 在此完成。
func Adapt[P any](fn TypedRun[P]) Runner {
	return runnerFunc(func(ctx context.Context, raw []byte, report ReportFunc) (any, error) {
		var p P
		if len(raw) > 0 {
			if err := json.Unmarshal(raw, &p); err != nil {
				return nil, fmt.Errorf("decode %T payload: %w", p, err)
			}
		}
		return fn(ctx, p, report)
	})
}

type runnerFunc func(ctx context.Context, payload []byte, report ReportFunc) (any, error)

func (f runnerFunc) Run(ctx context.Context, payload []byte, report ReportFunc) (any, error) {
	return f(ctx, payload, report)
}
