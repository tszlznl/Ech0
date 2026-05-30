// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { useAuthStore } from '@/stores/auth'
import { i18n } from '@/locales'

/**
 * 解析 baseURL（与 service/request 的代理逻辑保持一致），用于原生 fetch 读取 SSE。
 * 不走 ofetch 是因为需要逐块读取 ReadableStream 做流式渲染。
 */
function resolveChatURL(): string {
  const base = import.meta.env.VITE_SERVICE_BASE_URL ?? ''
  let url = `${base}/chat`
  if (import.meta.env.VITE_PROXY === 'YES') {
    const proxyUrl = import.meta.env.VITE_PROXY_URL
    if (proxyUrl) {
      url = `${proxyUrl}/chat`
    }
  }
  return url
}

interface ChatStreamHandlers {
  onSources?: (sources: App.Api.Chat.ChatSource[]) => void
  onDelta?: (text: string) => void
  onError?: (message: string) => void
  onDone?: () => void
}

/**
 * 发起 Chat 流式问答（SSE）。用原生 fetch + ReadableStream 读取，
 * 以便携带 Authorization 头（EventSource 不支持自定义头）。
 * 返回一个 abort 函数用于中断。
 */
export function chatStream(question: string, handlers: ChatStreamHandlers): () => void {
  const authStore = useAuthStore()
  const controller = new AbortController()
  const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC'

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    'X-Timezone': timezone,
    'X-Locale': i18n.global.locale.value,
  }
  if (authStore.authHeader) {
    headers['Authorization'] = authStore.authHeader
  }

  const run = async () => {
    let resp: Response
    try {
      resp = await fetch(resolveChatURL(), {
        method: 'POST',
        headers,
        body: JSON.stringify({ question }),
        credentials: 'include',
        signal: controller.signal,
      })
    } catch (e) {
      if (!controller.signal.aborted) {
        handlers.onError?.(String(e))
      }
      return
    }

    if (!resp.ok || !resp.body) {
      handlers.onError?.(`HTTP ${resp.status}`)
      return
    }

    const reader = resp.body.getReader()
    const decoder = new TextDecoder()
    let buffer = ''

    // 解析 SSE：事件以空行分隔，每块含 "event:" 与 "data:" 行。
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

      switch (eventName) {
        case 'sources':
          handlers.onSources?.(payload as App.Api.Chat.ChatSource[])
          break
        case 'delta':
          handlers.onDelta?.((payload as { text: string }).text)
          break
        case 'error':
          handlers.onError?.((payload as { message: string }).message)
          break
        case 'done':
          handlers.onDone?.()
          break
      }
    }

    try {
      for (;;) {
        const { done, value } = await reader.read()
        if (done) break
        buffer += decoder.decode(value, { stream: true })

        let sepIndex: number
        // 事件之间用空行（\n\n）分隔
        while ((sepIndex = buffer.indexOf('\n\n')) !== -1) {
          const rawEvent = buffer.slice(0, sepIndex)
          buffer = buffer.slice(sepIndex + 2)
          if (rawEvent.trim().length > 0 && !rawEvent.startsWith(':')) {
            dispatch(rawEvent)
          }
        }
      }
      handlers.onDone?.()
    } catch (e) {
      if (!controller.signal.aborted) {
        handlers.onError?.(String(e))
      }
    }
  }

  run()

  return () => controller.abort()
}
