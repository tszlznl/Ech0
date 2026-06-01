// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

/**
 * 上下文窗口等「token 数量」的人类友好单位解析 / 回显。
 * 约定：k = 1000、m = 1000000（与模型厂商标称窗口口径一致，如 256k / 1m）。
 */

const K = 1000
const M = 1000 * 1000

/**
 * 把用户输入（如 "256k" / "1m" / "128000" / "1.5m"）解析成 token 数。
 * 空串或无法识别一律返回 0（视为「未配置」，由后端走保守默认）。
 */
export function parseTokenSize(input: string): number {
  const s = (input ?? '').trim().toLowerCase()
  if (s === '') return 0
  const match = s.match(/^(\d+(?:\.\d+)?)\s*([km])?$/)
  if (!match) return 0
  const value = Number.parseFloat(match[1])
  if (!Number.isFinite(value)) return 0
  const unit = match[2]
  const tokens = unit === 'm' ? value * M : unit === 'k' ? value * K : value
  return Math.max(0, Math.round(tokens))
}

/**
 * 把 token 数回显成最紧凑的 k/m 形式（256000 → "256k"、1000000 → "1m"）。
 * 0 返回空串，让输入框显示占位提示。
 */
export function formatTokenSize(tokens: number): string {
  if (!tokens || tokens <= 0) return ''
  if (tokens % M === 0) return `${tokens / M}m`
  if (tokens % K === 0) return `${tokens / K}k`
  return String(tokens)
}
