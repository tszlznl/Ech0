// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package helpers

import (
	"context"
	"testing"

	"github.com/lin-snow/ech0/pkg/busen"
)

// NewTestBus 返回一个用于测试的进程内事件总线，并在测试结束时关闭。
// 供 event/bus 测试与需要 live bus 的 service 测试（如 echo.PostEcho 会 Notify 事件）复用。
func NewTestBus(t *testing.T) *busen.Bus {
	t.Helper()
	b := busen.New()
	t.Cleanup(func() { _ = b.Close(context.Background()) })
	return b
}
