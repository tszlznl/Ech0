// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package i18n

import "encoding/json"

func unmarshalJSON(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
