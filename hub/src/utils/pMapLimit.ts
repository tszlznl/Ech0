// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

/**
 * 受限并发版的 `Promise.allSettled`：保留入参顺序、不因单项失败拒绝整体。
 *
 * Hub 聚合的所有跨域扇出（probe / connect / page-fetch）都共享这个上限；
 * 在实例数膨胀到 100 时避免一次性 100 路并发跨域请求带来的 DNS 风暴、
 * TLS 握手抢占与浏览器队列阻塞。
 */
export async function pMapLimit<T, R>(
  items: readonly T[],
  limit: number,
  fn: (item: T, index: number) => Promise<R>,
  options?: { onSettled?: (result: PromiseSettledResult<R>, index: number) => void },
): Promise<PromiseSettledResult<R>[]> {
  const results: PromiseSettledResult<R>[] = new Array(items.length)
  let next = 0
  const onSettled = options?.onSettled

  const worker = async () => {
    while (true) {
      const i = next++
      if (i >= items.length) return
      let result: PromiseSettledResult<R>
      try {
        result = { status: 'fulfilled', value: await fn(items[i]!, i) }
      } catch (reason) {
        result = { status: 'rejected', reason }
      }
      results[i] = result
      onSettled?.(result, i)
    }
  }

  const n = Math.max(1, Math.min(limit, items.length))
  await Promise.all(Array.from({ length: n }, worker))
  return results
}

/**
 * Hub 聚合扇出的并发上限。8 是在「实例多时不让浏览器/网络抢占」与
 * 「实例少时还能接近全并发」之间的折中（100 实例 8 并发 ≈ 13 个 batch）。
 */
export const HUB_FAN_OUT_LIMIT = 8
