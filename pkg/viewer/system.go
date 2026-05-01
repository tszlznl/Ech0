// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package viewer

// NewSystemViewer returns a system-scoped viewer.
// For current simplified model, system and anonymous share the same behavior.
func NewSystemViewer() *NoopViewer { return NewNoopViewer() }
