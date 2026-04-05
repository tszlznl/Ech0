/**
 * 与 web/src/stores/hub.ts 对齐的多源归并：
 * 每实例独立缓冲池 + 按 createdTs 全局多路归并取数（batchSize 条/批）。
 * 请求使用 POST /api/echo/query（对应 web 中已 deprecated 的 echo/page 分页语义）。
 */
import { ref } from 'vue'
import type { HubInstance } from '../types/hub'
import type { EchoPost } from '../types/echo'
import { queryInstancePage } from '../services/echoApi'
import { normalizeHubInstanceUrl } from '../utils/hubUrl'
import { resolveHubInstanceLogo } from '../utils/resolveHubLogoUrl'

/** Hub 聚合不展示带 Extension 的 Echo（音乐/视频/GitHub/网站卡片等） */
function hasNoExtension(post: EchoPost): boolean {
  return post.extension == null
}

const PAGE_SIZE = 10
const BATCH_SIZE = 10

interface HubBufferState {
  url: string
  instanceId: string
  buffer: App.Api.Hub.Echo[]
  currentPage: number
  hasMore: boolean
  isLoading: boolean
}

function toHubEcho(
  post: EchoPost,
  inst: HubInstance,
  logos: Map<string, string>,
): App.Api.Hub.Echo {
  const instKey = normalizeHubInstanceUrl(inst.url)
  const rawLogo = logos.get(instKey) ?? ''
  const logo = resolveHubInstanceLogo(rawLogo, instKey)
  return {
    ...post,
    username: post.username ?? '',
    private: post.private ?? false,
    user_id: post.user_id ?? '',
    fav_count: post.fav_count ?? 0,
    tags: post.tags ?? [],
    createdTs: new Date(post.created_at).getTime(),
    server_name: inst.id,
    server_url: instKey,
    virtual_key: `${instKey}-${post.id}`,
    logo,
  }
}

export function useHubMergeFeed() {
  const hubStates = ref<Map<string, HubBufferState>>(new Map())
  const echoList = ref<App.Api.Hub.Echo[]>([])
  const existingIds = ref<Set<string>>(new Set())

  const isPreparing = ref(true)
  const isLoading = ref(false)
  const hasTriedInitialLoad = ref(false)
  const hasMore = ref(true)

  const fetchErrors = ref<{ instance: HubInstance; message: string }[]>([])
  /** 规范化实例 URL → `/api/connect` 返回的 logo 字段（在 toHubEcho 内再 resolve 为绝对 URL） */
  const instanceLogosByUrl = ref<Map<string, string>>(new Map())

  function setInstanceLogos(map: Map<string, string>) {
    instanceLogosByUrl.value = map
  }

  function reset() {
    hubStates.value.clear()
    echoList.value = []
    existingIds.value.clear()
    instanceLogosByUrl.value = new Map()
    isPreparing.value = true
    isLoading.value = false
    hasTriedInitialLoad.value = false
    hasMore.value = true
    fetchErrors.value = []
  }

  async function fetchHubPage(hubUrl: string, signal?: AbortSignal) {
    const state = hubStates.value.get(hubUrl)
    if (!state || state.isLoading || !state.hasMore) return

    state.isLoading = true
    try {
      const items = await queryInstancePage(
        hubUrl,
        {
          page: state.currentPage,
          pageSize: PAGE_SIZE,
          search: '',
          tagIds: [],
          sortBy: '',
          sortOrder: 'desc',
        },
        signal,
      )

      const inst: HubInstance = { id: state.instanceId, url: state.url }
      const mapped = items
        .filter(hasNoExtension)
        .map((echo) => toHubEcho(echo, inst, instanceLogosByUrl.value))

      mapped.sort((a, b) => b.createdTs - a.createdTs)
      state.buffer.push(...mapped)
      state.currentPage++
      state.hasMore = items.length >= PAGE_SIZE
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err)
      const state = hubStates.value.get(hubUrl)
      if (state) {
        fetchErrors.value.push({ instance: { id: state.instanceId, url: hubUrl }, message })
        state.hasMore = false
      }
    } finally {
      const s = hubStates.value.get(hubUrl)
      if (s) s.isLoading = false
    }
  }

  /**
   * 初始化各实例状态并并行拉取第一页填入缓冲池（与 web getHubInfoList 末尾一致）。
   */
  async function prepareInstances(instances: HubInstance[], signal?: AbortSignal) {
    isPreparing.value = true
    hasTriedInitialLoad.value = false
    hubStates.value.clear()
    fetchErrors.value = []

    for (const inst of instances) {
      const urlKey = normalizeHubInstanceUrl(inst.url)
      hubStates.value.set(urlKey, {
        url: urlKey,
        instanceId: inst.id,
        buffer: [],
        currentPage: 1,
        hasMore: true,
        isLoading: false,
      })
    }

    await Promise.all([...hubStates.value.keys()].map((url) => fetchHubPage(url, signal)))
    isPreparing.value = false
  }

  async function loadEchoListPage(signal?: AbortSignal) {
    if (isLoading.value || isPreparing.value) return

    const canLoadMore = [...hubStates.value.values()].some((s) => s.hasMore || s.buffer.length > 0)
    if (!canLoadMore) {
      hasMore.value = false
      hasTriedInitialLoad.value = true
      return
    }

    isLoading.value = true
    try {
      const result: App.Api.Hub.Echo[] = []
      let attempts = 0
      const maxAttempts = BATCH_SIZE * 3

      while (result.length < BATCH_SIZE && attempts < maxAttempts) {
        attempts++

        let maxTs = -1
        let maxHubUrl: string | null = null

        for (const [url, state] of hubStates.value) {
          const head = state.buffer[0]
          if (head) {
            const headTs = head.createdTs
            if (headTs > maxTs) {
              maxTs = headTs
              maxHubUrl = url
            }
          }
        }

        if (maxHubUrl === null) {
          const emptyHubsWithMore = [...hubStates.value.values()].filter(
            (s) => s.hasMore && !s.isLoading && s.buffer.length === 0,
          )

          if (emptyHubsWithMore.length === 0) break

          await Promise.all(emptyHubsWithMore.map((s) => fetchHubPage(s.url, signal)))
          continue
        }

        const state = hubStates.value.get(maxHubUrl)!
        const echo = state.buffer.shift()!

        const key = `${echo.server_url}-${echo.id}`
        if (!existingIds.value.has(key)) {
          existingIds.value.add(key)
          result.push(echo)
        }

        if (state.buffer.length < 3 && state.hasMore && !state.isLoading) {
          void fetchHubPage(maxHubUrl, signal)
        }
      }

      echoList.value.push(...result)

      hasMore.value = [...hubStates.value.values()].some((s) => s.hasMore || s.buffer.length > 0)
      hasTriedInitialLoad.value = true
    } finally {
      isLoading.value = false
    }
  }

  return {
    echoList,
    hubStates,
    isPreparing,
    isLoading,
    hasMore,
    hasTriedInitialLoad,
    fetchErrors,
    reset,
    setInstanceLogos,
    prepareInstances,
    loadEchoListPage,
  }
}
