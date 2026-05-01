// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package bus

import "reflect"

func safeTypeString(t reflect.Type) string {
	if t == nil {
		return ""
	}
	return t.String()
}
