import type { HubInstance } from '../types/hub'
import type { ApiResult } from '../types/echo'
import { normalizeHubInstanceUrl } from '../utils/hubUrl'

/** 与 web `GET {origin}/api/connect` 一致，用于取站点 logo（见 internal/handler/connect） */
export async function fetchInstanceConnect(
  instanceUrl: string,
  signal?: AbortSignal,
): Promise<App.Api.Connect.Connect | null> {
  const base = normalizeHubInstanceUrl(instanceUrl)
  const res = await fetch(`${base}/api/connect`, { credentials: 'omit', signal })
  if (!res.ok) return null
  const json: unknown = await res.json()
  if (typeof json !== 'object' || json === null) return null
  const r = json as ApiResult<App.Api.Connect.Connect>
  if (r.code !== 1 || !r.data) return null
  return r.data
}

export interface InstanceConnectSummary {
  urlKey: string
  id: string
  serverName: string
  username: string
  rawLogo: string
  /** Mirrors `GET /api/connect` `today_echos` (Connect widget heatmap dot). */
  todayEchos: number
}

/** 单次并行请求：logo Map（供聚合）+ 各实例摘要（供「活跃创作者」等） */
export async function fetchInstancesConnectBundle(
  instances: HubInstance[],
  signal?: AbortSignal,
): Promise<{ logos: Map<string, string>; summaries: InstanceConnectSummary[] }> {
  const logos = new Map<string, string>()
  const summaries: InstanceConnectSummary[] = []

  await Promise.all(
    instances.map(async (inst) => {
      const urlKey = normalizeHubInstanceUrl(inst.url)
      try {
        const data = await fetchInstanceConnect(urlKey, signal)
        if (!data) return
        const raw = data.logo?.trim() ?? ''
        if (raw) logos.set(urlKey, raw)
        summaries.push({
          urlKey,
          id: inst.id,
          serverName: data.server_name?.trim() || inst.id,
          username: data.sys_username?.trim() ?? '',
          rawLogo: raw,
          todayEchos: typeof data.today_echos === 'number' ? data.today_echos : 0,
        })
      } catch {
        /* 单实例失败不影响其它实例 */
      }
    }),
  )

  summaries.sort((a, b) => a.serverName.localeCompare(b.serverName, 'en'))
  return { logos, summaries }
}
