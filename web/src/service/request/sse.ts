// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { getApiUrl, buildCommonHeaders } from './shared'

export interface SSEStreamOptions {
  /** 相对路径，内部用 getApiUrl() 拼出完整后端地址（与全站请求同一套 baseURL/proxy 逻辑） */
  path: string
  /** 请求体，会被 JSON.stringify；GET 可省略 */
  body?: unknown
  method?: 'GET' | 'POST'
  /** 解析后的一帧事件：name 取自 `event:` 行，data 为 `data:` 行 JSON.parse 的结果 */
  onEvent: (name: string, data: unknown) => void
  /** 传输/网络错误（非中断），与服务端 `error` 事件分开由调用方各自处理 */
  onError?: (message: string) => void
  /** 流正常结束（reader 读完）时回调 */
  onClose?: () => void
}

/**
 * 统一的 SSE 传输：基于 fetch + ReadableStream（不用 EventSource——这样能带
 * Authorization 头、统一 abort、复用公共头）。收口 baseURL、公共头、AbortController、
 * `event:`/`data:`/`\n\n` 帧解析。返回 abort 函数用于中断。
 */
export function sseStream(opts: SSEStreamOptions): () => void {
  const controller = new AbortController()

  const dispatch = (rawEvent: string) => {
    let eventName = 'message'
    const dataLines: string[] = []
    for (const line of rawEvent.split('\n')) {
      if (line.startsWith('event:')) {
        eventName = line.slice(6).trim()
      } else if (line.startsWith('data:')) {
        dataLines.push(line.slice(5).trim())
      }
    }
    if (dataLines.length === 0) return
    let payload: unknown
    try {
      payload = JSON.parse(dataLines.join('\n'))
    } catch {
      return
    }
    opts.onEvent(eventName, payload)
  }

  const run = async () => {
    let resp: Response
    try {
      resp = await fetch(`${getApiUrl()}${opts.path}`, {
        method: opts.method ?? 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...buildCommonHeaders(),
        },
        body: opts.body !== undefined ? JSON.stringify(opts.body) : undefined,
        credentials: 'include',
        signal: controller.signal,
      })
    } catch (e) {
      if (!controller.signal.aborted) {
        opts.onError?.(String(e))
      }
      return
    }

    if (!resp.ok || !resp.body) {
      opts.onError?.(`HTTP ${resp.status}`)
      return
    }

    const reader = resp.body.getReader()
    const decoder = new TextDecoder()
    let buffer = ''

    try {
      for (;;) {
        const { done, value } = await reader.read()
        if (done) break
        buffer += decoder.decode(value, { stream: true })

        // 事件之间用空行（\n\n）分隔
        let sepIndex: number
        while ((sepIndex = buffer.indexOf('\n\n')) !== -1) {
          const rawEvent = buffer.slice(0, sepIndex)
          buffer = buffer.slice(sepIndex + 2)
          // 跳过注释帧（如 ": keep-alive"）
          if (rawEvent.trim().length > 0 && !rawEvent.startsWith(':')) {
            dispatch(rawEvent)
          }
        }
      }
      opts.onClose?.()
    } catch (e) {
      if (!controller.signal.aborted) {
        opts.onError?.(String(e))
      }
    }
  }

  run()

  return () => controller.abort()
}
