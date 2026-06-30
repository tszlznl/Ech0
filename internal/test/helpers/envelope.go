// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package helpers

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

// APIResult 是统一响应封套 commonModel.Result[T] 的测试侧投影（Data 保留为原始 JSON）。
// 字段对齐 internal/model/common.Result 的 json tag，便于断言 i18n 错误契约。
type APIResult struct {
	Code       int             `json:"code"`
	Msg        string          `json:"msg"`
	ErrorCode  string          `json:"error_code"`
	MessageKey string          `json:"message_key"`
	Data       json.RawMessage `json:"data"`
}

// ParseResult 解析 httptest 录制的响应体为 APIResult。
func ParseResult(t *testing.T, rec *httptest.ResponseRecorder) APIResult {
	t.Helper()
	var r APIResult
	if err := json.Unmarshal(rec.Body.Bytes(), &r); err != nil {
		t.Fatalf("helpers: parse result envelope: %v\nbody: %s", err, rec.Body.String())
	}
	return r
}

// DecodeData 把响应 Data 字段解码进 dest（指针）。
func DecodeData(t *testing.T, raw json.RawMessage, dest any) {
	t.Helper()
	if err := json.Unmarshal(raw, dest); err != nil {
		t.Fatalf("helpers: decode data: %v\nraw: %s", err, string(raw))
	}
}
