import type { ApiResult, EchoPost, EchoQueryPage } from '../types/echo'

export interface EchoQueryBody {
  page: number
  pageSize: number
  search: string
  tagIds: string[]
  sortBy: string
  sortOrder: string
}

const DEFAULT_QUERY: EchoQueryBody = {
  page: 1,
  pageSize: 10,
  search: '',
  tagIds: [],
  sortBy: '',
  sortOrder: 'desc',
}

/** 成功时 code === 1（见 internal/model/common/result.go DEFAULT_SUCCESS_CODE） */
export async function queryInstancePage(
  instanceUrl: string,
  body: Partial<EchoQueryBody> = {},
  signal?: AbortSignal,
): Promise<EchoPost[]> {
  const merged: EchoQueryBody = { ...DEFAULT_QUERY, ...body }
  const res = await fetch(`${instanceUrl}/api/echo/query`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      page: merged.page,
      pageSize: merged.pageSize,
      search: merged.search,
      tagIds: merged.tagIds,
      sortBy: merged.sortBy,
      sortOrder: merged.sortOrder,
    }),
    credentials: 'omit',
    signal,
  })

  if (!res.ok) {
    throw new Error(`HTTP ${res.status}`)
  }

  const json: unknown = await res.json()
  if (typeof json !== 'object' || json === null) {
    throw new Error('invalid JSON')
  }
  const r = json as ApiResult<EchoQueryPage>
  if (r.code !== 1 || !r.data || !Array.isArray(r.data.items)) {
    throw new Error(r.msg || 'query failed')
  }

  return r.data.items
}
