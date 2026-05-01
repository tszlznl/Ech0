// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package util

import "time"

const (
	DefaultTimezoneHeader = "X-Timezone"
	defaultTimezone       = "UTC"
)

// NormalizeTimezone 返回可用且规范的时区名称，非法输入回退为 UTC。
func NormalizeTimezone(tz string) string {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return defaultTimezone
	}
	return loc.String()
}

// LoadLocationOrUTC 加载时区，非法输入回退为 UTC。
func LoadLocationOrUTC(tz string) *time.Location {
	normalized := NormalizeTimezone(tz)
	loc, err := time.LoadLocation(normalized)
	if err != nil {
		return time.UTC
	}
	return loc
}
