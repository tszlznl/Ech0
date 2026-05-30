// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { request } from '../request'

// 获取 Embedding 向量设置
export function fetchGetEmbeddingSettings() {
  return request<App.Api.Embedding.EmbeddingSetting>({
    url: '/embedding/settings',
    method: 'GET',
  })
}

// 更新 Embedding 向量设置
export function fetchUpdateEmbeddingSettings(data: App.Api.Embedding.EmbeddingSettingDto) {
  return request<null>({
    url: '/embedding/settings',
    method: 'PUT',
    data,
  })
}

// 触发全量重建向量索引
export function fetchReindexEmbeddings() {
  return request<App.Api.Embedding.ReindexResult>({
    url: '/embedding/reindex',
    method: 'POST',
  })
}
