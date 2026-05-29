// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Chat / Embedding 相关类型（通过命名空间合并扩展 App.Api）。
declare namespace App {
  namespace Api {
    namespace Chat {
      // Embedding 向量设置
      type EmbeddingSetting = {
        enable: boolean
        model: string
        api_key: string
        base_url: string
        dim: number
      }

      type EmbeddingSettingDto = EmbeddingSetting

      // 回填索引结果
      type ReindexResult = {
        total: number
        indexed: number
        skipped: number
        failed: number
      }

      // 检索命中的来源 Echo
      type ChatSource = {
        echo_id: string
        content: string
        username: string
        echo_created: number
        distance: number
      }

      // 一条聊天消息（前端会话内）
      type ChatMessage = {
        role: 'user' | 'assistant'
        content: string
        sources?: ChatSource[]
      }

      // SSE 事件载荷
      type StreamEvent =
        | { type: 'sources'; data: ChatSource[] }
        | { type: 'delta'; data: { text: string } }
        | { type: 'error'; data: { message: string } }
        | { type: 'done'; data: { done: boolean } }
    }
  }
}
