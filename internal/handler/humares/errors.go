// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package humares

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	i18nUtil "github.com/lin-snow/ech0/internal/i18n"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	errUtil "github.com/lin-snow/ech0/internal/util/err"
)

// ErrorBody 仅用于 OpenAPI 文档：与错误响应的 wire 形态（Result[any]）一致，
// 但有干净的 schema 名（避免 Result[any] 被命名为 "ResultInterface {}"）。
type ErrorBody struct {
	Code          int            `json:"code" doc:"业务状态码，失败为 0"`
	Message       string         `json:"msg" doc:"状态描述（回退文案）"`
	ErrorCode     string         `json:"error_code,omitempty" doc:"稳定的业务错误码"`
	MessageKey    string         `json:"message_key,omitempty" doc:"i18n 翻译 key"`
	MessageParams map[string]any `json:"message_params,omitempty" doc:"i18n 模板参数"`
	Data          any            `json:"data" doc:"失败时为 null"`
}

// apiError 是 Huma 的自定义错误体。实现 huma.StatusError，且序列化为统一的 Result 信封——
// Huma 在 handler 返回 error 时会把该值原样写为响应体（与 response.Execute 错误路径一致）。
type apiError struct {
	status int
	body   commonModel.Result[any]
}

func (e *apiError) Error() string  { return e.body.Message }
func (e *apiError) GetStatus() int { return e.status }

// MarshalJSON 让错误响应体就是 Result 信封本身（而非 huma 默认的 problem+json）。
func (e *apiError) MarshalJSON() ([]byte, error) { return json.Marshal(e.body) }

// Schema 实现 huma.SchemaProvider：apiError 的字段未导出（靠 MarshalJSON 出 wire），
// 若不提供 schema，OpenAPI 里错误响应会是空对象。这里让它文档化为 Result 信封 schema。
func (e *apiError) Schema(r huma.Registry) *huma.Schema {
	return r.Schema(reflect.TypeFor[ErrorBody](), true, "ErrorBody")
}

// Err 把业务 error 转成 Huma 错误，复刻 response.Execute 的 BizError → message 优先级：
//  1. *BizError：取 MessageKey，缺失时按 Code 映射；带 Params。
//  2. 其余 error：按消息文本映射 message_key。
//
// 业务错误统一 HTTP 400（与现有 gin 行为一致）；鉴权 401/403 由 Bridge 包裹的中间件直接写出。
func Err(ctx context.Context, err error) error {
	base := errUtil.HandleError(&commonModel.ServerError{Err: err})
	loc := localizerFrom(ctx)

	var bizErr *commonModel.BizError
	if errors.As(err, &bizErr) {
		key := strings.TrimSpace(bizErr.MessageKey)
		if key == "" {
			key = commonModel.MessageKeyFromErrorCode(bizErr.Code)
		}
		msg := i18nUtil.Localize(loc, key, base, bizErr.Params)
		return &apiError{
			status: http.StatusBadRequest,
			body:   commonModel.FailWithLocalized[any](msg, bizErr.Code, key, bizErr.Params),
		}
	}

	if key := commonModel.MessageKeyFromMessage(base); key != "" {
		msg := i18nUtil.Localize(loc, key, base, nil)
		return &apiError{
			status: http.StatusBadRequest,
			body:   commonModel.FailWithLocalized[any](msg, "", key, nil),
		}
	}

	return &apiError{status: http.StatusBadRequest, body: commonModel.Fail[any](base)}
}

// installErrorModel 覆写 Huma 的全局错误构造器，让框架级错误（请求体解析失败、
// 路径/查询参数校验失败等）也统一成 Result 信封而非 problem+json。由 NewAPI 调用。
func installErrorModel() {
	huma.NewErrorWithContext = func(hctx huma.Context, status int, msg string, _ ...error) huma.StatusError {
		loc := i18nUtil.LocalizerFromGin(humagin.Unwrap(hctx))
		localized := i18nUtil.Localize(loc, commonModel.MsgKeyInvalidQueryParams, msg, nil)
		return &apiError{
			status: status,
			body:   commonModel.FailWithLocalized[any](localized, commonModel.ErrCodeInvalidRequest, commonModel.MsgKeyInvalidQueryParams, nil),
		}
	}
	huma.NewError = func(status int, msg string, _ ...error) huma.StatusError {
		return &apiError{
			status: status,
			body:   commonModel.FailWithLocalized[any](msg, commonModel.ErrCodeInvalidRequest, commonModel.MsgKeyInvalidQueryParams, nil),
		}
	}
}
