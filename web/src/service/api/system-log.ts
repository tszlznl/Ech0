// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { request } from '../request'

export function fetchSystemLogs(params: App.Api.SystemLog.QueryParams) {
  const query = new URLSearchParams()
  query.append('tail', String(params.tail ?? 200))

  const level = params.level?.trim()
  if (level && level !== 'all') {
    query.append('level', level)
  }

  const keyword = params.keyword?.trim()
  if (keyword) {
    query.append('keyword', keyword)
  }

  return request<App.Api.SystemLog.Entry[]>({
    url: `/system/logs?${query.toString()}`,
    method: 'GET',
  })
}
