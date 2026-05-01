// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { normalizeHubInstanceUrl } from './hubUrl'

/**
 * 将 `/api/connect` 返回的 logo 转为可在 Hub 页面直接用于 `<img src>` 的绝对地址。
 * 主站 `resolveAvatarUrl` 依赖当前站点的 VITE_SERVICE_BASE_URL；Hub 跨域访问各实例，应对齐到**该实例 origin**。
 */
export function resolveHubInstanceLogo(rawUrl: string | undefined, instanceOrigin: string): string {
  const value = (rawUrl || '').trim()
  if (!value) return ''

  if (/^https?:\/\//i.test(value)) return value

  const base = normalizeHubInstanceUrl(instanceOrigin)
  if (value.startsWith('/')) return `${base}${value}`
  return `${base}/${value}`
}
