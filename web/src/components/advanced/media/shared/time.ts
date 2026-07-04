// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

/**
 * 把秒数格式化为播放器时间戳：不足一小时用 m:ss，超过一小时用 h:mm:ss。
 * 非法/未知（NaN、负数、未加载出时长）统一回落到 0:00，避免控件里出现 "NaN:NaN"。
 */
export function formatMediaTime(seconds: number | null | undefined): string {
  if (seconds == null || !Number.isFinite(seconds) || seconds < 0) return '0:00'

  const total = Math.floor(seconds)
  const hours = Math.floor(total / 3600)
  const minutes = Math.floor((total % 3600) / 60)
  const secs = total % 60
  const paddedSecs = String(secs).padStart(2, '0')

  if (hours > 0) {
    return `${hours}:${String(minutes).padStart(2, '0')}:${paddedSecs}`
  }
  return `${minutes}:${paddedSecs}`
}
