// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { request } from '@/service/request'
import { sseStream } from '@/service/request/sse'

/** 获取当前用户的持久化 Chat 会话（重载页面恢复展示用） */
export function getChatSession() {
  return request<App.Api.Chat.ChatMessage[]>({
    url: `/chat/session`,
    method: 'GET',
  })
}

/** 清除当前用户的持久化 Chat 会话（不可恢复） */
export function clearChatSession() {
  return request({
    url: `/chat/session`,
    method: 'DELETE',
  })
}

interface ChatStreamHandlers {
  /** 模型决定检索时触发（Agent 形态，可多次），携带本次检索关键词 */
  onSearching?: (query: string) => void
  /** 命中来源到达（可多次增量），调用方需累积去重 */
  onSources?: (sources: App.Api.Chat.ChatSource[]) => void
  /** 区间聚合总结的覆盖度到达（summarize_echos） */
  onCoverage?: (coverage: App.Api.Chat.ChatCoverage) => void
  onDelta?: (text: string) => void
  onError?: (message: string) => void
  onDone?: () => void
}

/**
 * 发起 Chat 流式问答（SSE）。传输细节（fetch + ReadableStream、公共头、abort、帧解析）
 * 收口在 service/request/sse.ts；此处仅把语义事件映射为类型化 handler。
 * 返回一个 abort 函数用于中断。
 */
export function chatStream(question: string, handlers: ChatStreamHandlers): () => void {
  let done = false
  const finish = () => {
    if (done) return
    done = true
    handlers.onDone?.()
  }

  return sseStream({
    path: '/chat',
    body: { question },
    onEvent: (event, data) => {
      switch (event) {
        case 'searching':
          handlers.onSearching?.((data as { name: string; query: string }).query)
          break
        case 'sources':
          handlers.onSources?.(data as App.Api.Chat.ChatSource[])
          break
        case 'coverage':
          handlers.onCoverage?.(data as App.Api.Chat.ChatCoverage)
          break
        case 'delta':
          handlers.onDelta?.((data as { text: string }).text)
          break
        case 'error':
          handlers.onError?.((data as { message: string }).message)
          break
        case 'done':
          finish()
          break
      }
    },
    onError: (message) => handlers.onError?.(message),
    onClose: finish,
  })
}
