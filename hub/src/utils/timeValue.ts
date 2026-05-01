// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

/**
 * 将「存储层时间」统一为毫秒时间戳。
 *
 * 兼容：
 * - 新版：Unix 时间戳（number，秒或毫秒；数值小于 1e12 时按秒处理）
 * - 旧版：ISO 8601 等可被 `Date.parse` 解析的文本
 *
 * 无法解析时返回 `0`（便于排序/归并时落在末尾而不抛错）。
 */
export function timeValueToMs(raw: unknown): number {
  if (raw == null) return 0
  if (typeof raw === 'number') {
    if (!Number.isFinite(raw)) return 0
    return raw < 1e12 ? raw * 1000 : raw
  }
  if (typeof raw === 'string') {
    const s = raw.trim()
    if (!s) return 0
    if (/^-?\d+(\.\d+)?$/.test(s)) {
      const n = Number(s)
      if (!Number.isFinite(n)) return 0
      return n < 1e12 ? n * 1000 : n
    }
    const ms = Date.parse(s)
    return Number.isNaN(ms) ? 0 : ms
  }
  return 0
}

/** 与后端 `int64` Unix 秒对齐；由 `timeValueToMs` 向下取整。 */
export function timeValueToUnixSeconds(raw: unknown): number {
  const ms = timeValueToMs(raw)
  if (ms <= 0) return 0
  return Math.floor(ms / 1000)
}
