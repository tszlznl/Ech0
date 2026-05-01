// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package viewer provides a unified viewer context abstraction.
package viewer

// Context defines the current viewer identity.
type Context interface {
	UserID() string
	TokenType() string
	Scopes() []string
	Audience() []string
	TokenID() string
}
