// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package transaction

import "github.com/google/wire"

var TransactorSet = wire.NewSet(
	NewGormTransactor,
	wire.Bind(new(Transactor), new(*GormTransactor)),
)

var ProviderSet = TransactorSet
