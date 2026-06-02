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

// 提交全量重建向量索引作业（异步：起即返回作业状态，前端轮询进度）
export function fetchReindexEmbeddings() {
  return request<App.Api.Embedding.ReindexStatus>({
    url: '/embedding/reindex',
    method: 'POST',
  })
}

// 查询重建索引作业状态（按 type 轮询，无需 id）
export function fetchReindexStatus() {
  return request<App.Api.Embedding.ReindexStatus>({
    url: '/embedding/reindex/status',
    method: 'GET',
  })
}

// 取消进行中的重建索引作业
export function fetchCancelReindex() {
  return request<App.Api.Embedding.ReindexStatus>({
    url: '/embedding/reindex/cancel',
    method: 'POST',
  })
}
