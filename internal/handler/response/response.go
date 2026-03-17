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

// Response 代表 handler 层的执行结果封装
//
// swagger:model Response
type Response struct {
	// Code 状态码，非0时表示自定义HTTP业务状态码
	Code int `json:"code"`

	// Data 响应数据，具体内容因接口而异
	Data any `json:"data,omitempty"`

	// Msg 返回信息，通常是状态描述
	Msg string `json:"msg"`

	// ErrorCode 业务错误码，可选
	ErrorCode string `json:"error_code,omitempty"`

	// MessageKey 国际化消息 key，可选
	MessageKey string `json:"message_key,omitempty"`

	// MessageParams 国际化模板参数，可选
	MessageParams map[string]any `json:"message_params,omitempty"`

	// Err 错误信息，序列化时忽略（仅供内部日志使用）
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

		// 支持自定义 code
		if res.Code != 0 {
			ctx.JSON(http.StatusOK, commonModel.OKWithCode(res.Data, res.Code, successMsg))
		} else {
			ctx.JSON(http.StatusOK, commonModel.OK(res.Data, successMsg))
		}
	}
}
