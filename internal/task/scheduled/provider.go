// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package scheduled

import "github.com/google/wire"

// ProviderSet 提供各领域的定时 Task。装配进 task.Manager 由 di.ProvideTaskManager 完成。
var ProviderSet = wire.NewSet(
	NewCleanup,
	NewDeadLetter,
	NewBackup,
	NewVisitorSnapshot,
)
