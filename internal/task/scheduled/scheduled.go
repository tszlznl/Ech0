// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package scheduled 实现各领域的定时 Task，连接通用 task 框架与领域 service。
// 对应 job/runner 之于 job：task 核心包保持纯净（只依赖 gocron + 标准库），
// 领域依赖收在本子包，故无 import 环。
package scheduled

// logModule 是本子包统一的日志 module 字段值，与 task 核心包对齐。
const logModule = "task"
