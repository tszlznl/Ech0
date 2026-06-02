// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package runner

import "github.com/google/wire"

// ProviderSet 提供各领域 Runner。装配进 job.Manager 由 di.ProvideJobManager 完成。
var ProviderSet = wire.NewSet(
	NewReindexRunner,
	NewMigrationRunner,
)
