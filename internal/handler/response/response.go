// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	i18nUtil "github.com/lin-snow/ech0/internal/i18n"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	errorUtil "github.com/lin-snow/ech0/internal/util/err"
)

// swagger:model Response
type Response struct {
	Code int `json:"code"`

	Data any `json:"data,omitempty"`

	Msg string `json:"msg"`

	ErrorCode string `json:"error_code,omitempty"`

	MessageKey string `json:"message_key,omitempty"`

	MessageParams map[string]any `json:"message_params,omitempty"`

	// swagger:ignore
	Err error `json:"-"`
}

// Execute 包装器，自动根据 Response 返回统一格式的 HTTP 响应 (仅处理返回类型为JSON的handler)
func Execute(fn func(ctx *gin.Context) Response) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		res := fn(ctx)
		if res.Err != nil {
			msg := errorUtil.HandleError(&commonModel.ServerError{
				Msg: res.Msg,
				Err: res.Err,
			})
			localizer := i18nUtil.LocalizerFromGin(ctx)

			var bizErr *commonModel.BizError
			if errors.As(res.Err, &bizErr) {
				messageKey := strings.TrimSpace(bizErr.MessageKey)
				if messageKey == "" {
					messageKey = commonModel.MessageKeyFromErrorCode(bizErr.Code)
				}
				msg = i18nUtil.Localize(localizer, messageKey, msg, bizErr.Params)
				ctx.JSON(http.StatusBadRequest, commonModel.FailWithLocalized[string](msg, bizErr.Code, messageKey, bizErr.Params))
				return
			}

			if res.ErrorCode != "" {
				messageKey := strings.TrimSpace(res.MessageKey)
				if messageKey == "" {
					messageKey = commonModel.MessageKeyFromErrorCode(res.ErrorCode)
				}
				msg = i18nUtil.Localize(localizer, messageKey, msg, res.MessageParams)
				ctx.JSON(http.StatusBadRequest, commonModel.FailWithLocalized[string](msg, res.ErrorCode, messageKey, res.MessageParams))
				return
			}

			messageKey := commonModel.MessageKeyFromMessage(msg)
			msg = i18nUtil.Localize(localizer, messageKey, msg, nil)
			if messageKey != "" {
				ctx.JSON(http.StatusBadRequest, commonModel.FailWithLocalized[string](msg, "", messageKey, nil))
				return
			}

			ctx.JSON(http.StatusBadRequest, commonModel.Fail[string](msg))
			return
		}

		successMsg := res.Msg
		messageKey := strings.TrimSpace(res.MessageKey)
		if messageKey == "" {
			messageKey = commonModel.MessageKeyFromMessage(res.Msg)
		}
		if messageKey != "" {
			successMsg = i18nUtil.Localize(i18nUtil.LocalizerFromGin(ctx), messageKey, res.Msg, res.MessageParams)
		}

		if res.Code != 0 {
			ctx.JSON(http.StatusOK, commonModel.OKWithCode(res.Data, res.Code, successMsg))
		} else {
			ctx.JSON(http.StatusOK, commonModel.OK(res.Data, successMsg))
		}
	}
}
