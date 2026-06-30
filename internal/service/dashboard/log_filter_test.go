// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"testing"

	logUtil "github.com/lin-snow/ech0/internal/util/log"
)

// TestMatchesSystemLogFilter 白盒覆盖 matchesSystemLogFilter 的纯过滤逻辑：
// level 为空/"all" 跳过级别判断，否则按 lower(entry.Level)==level 比较；
// keyword 为空直接放行，否则对 lower(Msg+" "+Raw) 做子串匹配。
// 注意：生产调用方已对 level/keyword 做 ToLower+TrimSpace，故用例传入的均为小写。
func TestMatchesSystemLogFilter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		entry   logUtil.LogEntry
		level   string
		keyword string
		want    bool
	}{
		{
			name:    "level empty skips level check, empty keyword passes",
			entry:   logUtil.LogEntry{Level: "error", Msg: "boom"},
			level:   "",
			keyword: "",
			want:    true,
		},
		{
			name:    "level all skips level check, empty keyword passes",
			entry:   logUtil.LogEntry{Level: "info", Msg: "hello"},
			level:   "all",
			keyword: "",
			want:    true,
		},
		{
			name:    "level matches case-insensitively on entry.Level",
			entry:   logUtil.LogEntry{Level: "ERROR", Msg: "x"},
			level:   "error",
			keyword: "",
			want:    true,
		},
		{
			name:    "level mismatch rejects",
			entry:   logUtil.LogEntry{Level: "info", Msg: "x"},
			level:   "error",
			keyword: "",
			want:    false,
		},
		{
			name:    "level mismatch rejects even when keyword would match",
			entry:   logUtil.LogEntry{Level: "info", Msg: "needle"},
			level:   "error",
			keyword: "needle",
			want:    false,
		},
		{
			name:    "keyword empty passes after level passes",
			entry:   logUtil.LogEntry{Level: "info", Msg: "anything"},
			level:   "info",
			keyword: "",
			want:    true,
		},
		{
			name:    "keyword hits substring in Msg",
			entry:   logUtil.LogEntry{Level: "info", Msg: "connection refused"},
			level:   "",
			keyword: "refused",
			want:    true,
		},
		{
			name:    "keyword hits substring in Raw",
			entry:   logUtil.LogEntry{Level: "info", Msg: "ok", Raw: "trace id=abc123"},
			level:   "",
			keyword: "abc123",
			want:    true,
		},
		{
			name:    "keyword matches case-insensitively against Msg",
			entry:   logUtil.LogEntry{Level: "info", Msg: "Database Timeout"},
			level:   "",
			keyword: "timeout",
			want:    true,
		},
		{
			name:    "keyword spans Msg-Raw join boundary",
			entry:   logUtil.LogEntry{Level: "info", Msg: "foo", Raw: "bar"},
			level:   "",
			keyword: "foo bar",
			want:    true,
		},
		{
			name:    "keyword no match rejects",
			entry:   logUtil.LogEntry{Level: "info", Msg: "all good", Raw: "nothing here"},
			level:   "",
			keyword: "missing",
			want:    false,
		},
		{
			name:    "level and keyword both match",
			entry:   logUtil.LogEntry{Level: "WARN", Msg: "disk almost full", Raw: "/dev/sda"},
			level:   "warn",
			keyword: "almost",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := matchesSystemLogFilter(tt.entry, tt.level, tt.keyword); got != tt.want {
				t.Fatalf("matchesSystemLogFilter(%+v, %q, %q) = %v, want %v",
					tt.entry, tt.level, tt.keyword, got, tt.want)
			}
		})
	}
}
