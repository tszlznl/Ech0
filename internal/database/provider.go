// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package database

import (
	"sync"

	"github.com/google/wire"
	"gorm.io/gorm"
)

func ProvideDBProvider() func() *gorm.DB {
	var once sync.Once
	return func() *gorm.DB {
		once.Do(InitDatabase)
		return GetDB()
	}
}

var ProviderSet = wire.NewSet(ProvideDBProvider)
