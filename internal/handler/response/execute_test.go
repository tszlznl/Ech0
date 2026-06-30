// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	i18n "github.com/lin-snow/ech0/internal/i18n"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	"github.com/lin-snow/ech0/internal/test/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// runExecute 用 en-US localizer 驱动 Execute(fn) 一次，返回录制的响应。
// 挂 en-US localizer 是为了让 message_key 解析阶梯真正走到本地化分支。
func runExecute(t *testing.T, res Response) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Set(i18n.ContextLocalizerKey, i18n.NewLocalizer("en-US", ""))
	Execute(func(*gin.Context) Response { return res })(c)
	return rec
}

// ---------------------------------------------------------------------------
// 成功路径
// ---------------------------------------------------------------------------

func TestExecute_Success(t *testing.T) {
	cases := []struct {
		name        string
		res         Response
		wantCode    int
		wantMessage string
		wantData    string // 解码后的 Data（字符串负载）
	}{
		{
			name:        "code-zero-plain-data-empty-msg",
			res:         Response{Code: 0, Data: "payload", Msg: ""},
			wantCode:    commonModel.DEFAULT_SUCCESS_CODE,
			wantMessage: "",
			wantData:    "payload",
		},
		{
			name:        "code-zero-unmapped-msg-passthrough",
			res:         Response{Code: 0, Data: "payload", Msg: "custom hello"},
			wantCode:    commonModel.DEFAULT_SUCCESS_CODE,
			wantMessage: "custom hello",
			wantData:    "payload",
		},
		{
			name:        "code-zero-known-msg-localized",
			res:         Response{Code: 0, Data: "payload", Msg: commonModel.SUCCESS_MESSAGE},
			wantCode:    commonModel.DEFAULT_SUCCESS_CODE,
			wantMessage: "Request succeeded", // common.success @ en-US
			wantData:    "payload",
		},
		{
			name:        "code-zero-explicit-message-key-localized",
			res:         Response{Code: 0, Data: "payload", Msg: "raw", MessageKey: commonModel.MsgKeyCommonSuccess},
			wantCode:    commonModel.DEFAULT_SUCCESS_CODE,
			wantMessage: "Request succeeded",
			wantData:    "payload",
		},
		{
			name:        "nonzero-code-takes-okwithcode",
			res:         Response{Code: 7, Data: "payload", Msg: "custom"},
			wantCode:    7,
			wantMessage: "custom",
			wantData:    "payload",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := runExecute(t, tc.res)

			require.Equal(t, http.StatusOK, rec.Code)
			got := helpers.ParseResult(t, rec)
			assert.Equal(t, tc.wantCode, got.Code)
			assert.Equal(t, tc.wantMessage, got.Msg)
			// 成功封套不携带 error_code / message_key。
			assert.Empty(t, got.ErrorCode)
			assert.Empty(t, got.MessageKey)

			var data string
			helpers.DecodeData(t, got.Data, &data)
			assert.Equal(t, tc.wantData, data)
		})
	}
}

// ---------------------------------------------------------------------------
// 失败路径
// ---------------------------------------------------------------------------

func TestExecute_Failure(t *testing.T) {
	cases := []struct {
		name           string
		res            Response
		wantErrorCode  string
		wantMessageKey string
		wantMessage    string
	}{
		{
			name: "bizerror-carries-code-and-key",
			res: Response{
				Err: commonModel.NewBizErrorWithMessageKey(
					commonModel.ErrCodeInvalidQuery, "raw msg",
					commonModel.MsgKeyInvalidQueryParams, nil,
				),
			},
			wantErrorCode:  commonModel.ErrCodeInvalidQuery,
			wantMessageKey: commonModel.MsgKeyInvalidQueryParams,
			wantMessage:    "Invalid query parameters",
		},
		{
			name: "plain-error-msg-text-maps-to-key",
			res: Response{
				Err: errors.New("boom"),
				Msg: commonModel.AGENT_MODEL_MISSING,
			},
			wantErrorCode:  "", // 非 BizError 且未显式设 ErrorCode
			wantMessageKey: commonModel.MsgKeyAgentModelMissing,
			wantMessage:    "Agent model name is not configured or is empty",
		},
		{
			name: "explicit-errorcode-fallback-derives-key",
			res: Response{
				Err:       errors.New("boom"),
				Msg:       "totally unmapped text",
				ErrorCode: commonModel.ErrCodeInvalidQuery,
			},
			wantErrorCode:  commonModel.ErrCodeInvalidQuery,
			wantMessageKey: commonModel.MsgKeyInvalidQueryParams,
			wantMessage:    "Invalid query parameters",
		},
		{
			name: "no-code-no-key-plain-fail",
			res: Response{
				Err: errors.New("boom"),
				Msg: "totally unmapped text",
			},
			wantErrorCode:  "",
			wantMessageKey: "",
			wantMessage:    "totally unmapped text",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := runExecute(t, tc.res)

			// 失败统一 400，错误信息装在封套里。
			require.Equal(t, http.StatusBadRequest, rec.Code)
			got := helpers.ParseResult(t, rec)
			assert.Equal(t, commonModel.DEFAULT_FAILED_CODE, got.Code)
			assert.Equal(t, tc.wantErrorCode, got.ErrorCode)
			assert.Equal(t, tc.wantMessageKey, got.MessageKey)
			assert.Equal(t, tc.wantMessage, got.Msg)
		})
	}
}

// 未挂 localizer 时（ctx 无 localizer），message_key 仍被透出，
// 文本回退为默认文案（Localize 对 nil localizer 直接返回 defaultText）。
func TestExecute_Failure_NoLocalizerFallsBackToDefaultText(t *testing.T) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	Execute(func(*gin.Context) Response {
		return Response{
			Err: commonModel.NewBizErrorWithMessageKey(
				commonModel.ErrCodeInvalidQuery, "raw msg",
				commonModel.MsgKeyInvalidQueryParams, nil,
			),
		}
	})(c)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	got := helpers.ParseResult(t, rec)
	assert.Equal(t, commonModel.ErrCodeInvalidQuery, got.ErrorCode)
	assert.Equal(t, commonModel.MsgKeyInvalidQueryParams, got.MessageKey)
	assert.Equal(t, "raw msg", got.Msg) // 未本地化，回退默认文案
}
