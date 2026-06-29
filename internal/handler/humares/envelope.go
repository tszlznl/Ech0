// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package humares 为 Huma (type-first OpenAPI) 提供与现有 gin handler 一致的响应契约：
// 统一信封 Result[T]、i18n 错误本地化、以及复用现有 gin 鉴权中间件的桥接器。
// 它是 swaggo 注解迁移到 Huma 的共享基础设施，被 router 层用来注册 operation。
package humares

import (
	"context"

	i18nUtil "github.com/lin-snow/ech0/internal/i18n"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
)

// Envelope 是所有 JSON 端点的 Huma 输出包装。Huma 序列化其 Body 字段，
// 因此 OpenAPI 会从 Result[T] 反射出真实结构（data 不再是 any，而是具体 schema）。
//
// 线上 JSON 形态与现有 commonModel.Result[T] 完全一致：
// { code, msg, error_code?, message_key?, message_params?, data }，前端无需改动。
type Envelope[T any] struct {
	Body commonModel.Result[T]
}

// OK 构造成功响应。按 message 推导 message_key 并本地化 msg，与 response.Execute 成功路径一致；
// 前端仍可凭 message_key 自行渲染。localizer 由 injectLocalizer 注入到 ctx。
func OK[T any](ctx context.Context, data T, messages ...string) *Envelope[T] {
	return &Envelope[T]{Body: localizeResult(ctx, commonModel.OK(data, messages...))}
}

// OKWithCode 同 OK，但允许自定义业务 code（对应 response.Execute 的 res.Code != 0 分支）。
func OKWithCode[T any](ctx context.Context, data T, code int, messages ...string) *Envelope[T] {
	return &Envelope[T]{Body: localizeResult(ctx, commonModel.OKWithCode(data, code, messages...))}
}

// localizeResult 为成功信封补齐 message_key 并本地化 msg。
func localizeResult[T any](ctx context.Context, body commonModel.Result[T]) commonModel.Result[T] {
	body.MessageKey = commonModel.MessageKeyFromMessage(body.Message)
	if body.MessageKey != "" {
		body.Message = i18nUtil.Localize(localizerFrom(ctx), body.MessageKey, body.Message, body.MessageParams)
	}
	return body
}
