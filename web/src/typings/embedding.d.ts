// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Embedding 向量索引相关类型（通过命名空间合并扩展 App.Api）。
declare namespace App {
  namespace Api {
    namespace Embedding {
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
    }
  }
}
