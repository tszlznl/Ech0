// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package transaction

import (
	"context"

	"gorm.io/gorm"
)

func TxFromContext(ctx context.Context) (*gorm.DB, bool) {
	if ctx == nil {
		return nil, false
	}

	tx, ok := ctx.Value(TxKey).(*gorm.DB)
	return tx, ok
}

func HasTx(ctx context.Context) bool {
	_, ok := TxFromContext(ctx)
	return ok
}
