// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2024-2026 lin-snow

package template

import "embed"

// all: 包含以 _ / . 开头的文件名；Vite 8 会生成 _plugin-vue_export-helper-*.js，默认 embed 会排除此类文件。
//
//go:embed all:dist
var WebFS embed.FS
