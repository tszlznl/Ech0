// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

/**
 * 跨实例 fetch 的硬超时信号。没有它，单个吊死的实例会让首页/探索页等到
 * 浏览器默认超时（30s+）才放手；扇出到 100 个实例时这是首要瓶颈。
 *
 * 同时透传上层的 `outer` signal（用于路由切换/视图卸载时的级联取消）。
 */
export function timeoutSignal(outer: AbortSignal | undefined, timeoutMs: number): AbortSignal {
  const t = AbortSignal.timeout(timeoutMs)
  return outer ? AbortSignal.any([outer, t]) : t
}
