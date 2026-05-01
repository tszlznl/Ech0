// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package cache

import "github.com/google/wire"

func ProvideCache() (ICache[string, any], error) {
	return NewCache[string, any]()
}

var ProviderSet = wire.NewSet(ProvideCache)
