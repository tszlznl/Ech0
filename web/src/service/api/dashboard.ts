// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { request } from '../request'

export function fetchGetVisitorStats() {
  return request<App.Api.Dashboard.VisitorDayStat[]>({
    url: '/system/visitor-stats',
    method: 'GET',
  })
}
