import { defineStore } from 'pinia'
import { ref, computed, watch } from 'vue'
import { fetchQueryEchos, fetchGetTags } from '@/service/api'

export const useEchoStore = defineStore('echoStore', () => {
  /**
   * 将 Echo 的 id 归一化为去除首尾空格的字符串，
   * 用于 indexMap 的 key 一致性保证。
   */
  const normalizeEchoId = (echo: App.Api.Ech0.Echo): string => String(echo?.id ?? '').trim()

  /**
   * 将新到达的 Echo 列表合并进目标列表：
   * - 已存在（id 命中 indexMap）→ 原地更新
   * - 不存在 → 追加到末尾并记录索引
   * 保证列表不会出现重复项，同时维护 O(1) 的 id→index 查找。
   */
  const mergeEchoItems = (
    targetList: { value: App.Api.Ech0.Echo[] },
    targetIndexMap: { value: Map<string, number> },
    incoming: App.Api.Ech0.Echo[],
  ) => {
    incoming.forEach((item: App.Api.Ech0.Echo) => {
      const id = normalizeEchoId(item)
      if (!id) return
      const normalizedItem = { ...item, id }
      const idx = targetIndexMap.value.get(id)
      if (idx !== undefined) {
        targetList.value[idx] = normalizedItem
      } else {
        targetList.value.push(normalizedItem)
        targetIndexMap.value.set(id, targetList.value.length - 1)
      }
    })
  }

  // ─────────────────────────────────────────────
  //  统一查询状态（普通时间线与标签过滤共享同一套状态）
  // ─────────────────────────────────────────────

  const echoList = ref<App.Api.Ech0.Echo[]>([]) // 当前展示的 Echo 列表
  const echoIndexMap = ref(new Map<string, number>()) // id → echoList 下标，用于快速定位
  const isLoading = ref<boolean>(true) // 是否正在请求数据
  const total = ref<number>(0) // 当前查询条件下的总条数
  const pageSize = ref<number>(5) // 每页数量
  const page = ref<number>(0) // 已成功加载到的页码（0 表示尚未加载）
  const current = ref<number>(1) // 下一次要请求的页码（从 1 开始）
  const searchValue = ref<string>('') // 搜索关键词
  const hasMore = computed(() => total.value > echoList.value.length) // 是否还有更多数据
  const searchingMode = computed(() => searchValue.value.length > 0) // 是否处于搜索模式
  const echoToUpdate = ref<App.Api.Ech0.EchoToUpdate | null>(null) // 编辑中的 Echo 暂存

  const tagList = ref<App.Api.Ech0.Tag[]>([]) // 全部标签列表
  const tagOptions = computed<string[]>(() => tagList.value.map((tag) => tag.name))

  // ── 标签过滤模式 ──
  const isFilteringMode = ref<boolean>(false) // 是否正在按标签过滤
  const filteredTag = ref<App.Api.Ech0.Tag | null>(null) // 当前选中的过滤标签

  // ── 日期范围过滤（由高级搜索面板驱动；单位：Unix 秒） ──
  const dateFrom = ref<number | null>(null)
  const dateTo = ref<number | null>(null)
  const isDateRangeActive = computed(() => dateFrom.value !== null || dateTo.value !== null)

  // ── 多标签过滤（由高级搜索面板驱动）──
  // 与 filteredTag（单标签点击过滤）独立存在：
  //   - filteredTag：用户点击 Echo 卡片或标签页进入的“聚焦单标签”模式
  //   - selectedTagIds：高级搜索面板里可组合多个标签，后端以 OR 语义查询
  // buildQueryParams 会将二者合并去重传给后端。
  const selectedTagIds = ref<string[]>([])
  const isTagSelectionActive = computed(() => selectedTagIds.value.length > 0)

  // ─────────────────────────────────────────────
  //  watchers
  // ─────────────────────────────────────────────

  // 从搜索模式退出时（关键词清空），自动刷新列表回到默认时间线
  watch(searchingMode, (newValue, oldValue) => {
    if (newValue === false && oldValue === true) {
      refreshEchos()
    }
  })

  // ─────────────────────────────────────────────
  //  核心查询 actions
  // ─────────────────────────────────────────────

  /**
   * 根据当前状态构建统一查询参数。
   * - 普通模式：仅携带 page / pageSize / search
   * - 标签过滤模式：额外携带 tagIds
   */
  function buildQueryParams(): App.Api.Ech0.EchoQueryParams {
    const params: App.Api.Ech0.EchoQueryParams = {
      page: current.value,
      pageSize: pageSize.value,
      search: searchValue.value || undefined,
    }
    const tagIds = new Set<string>()
    if (isFilteringMode.value && filteredTag.value) {
      tagIds.add(filteredTag.value.id)
    }
    selectedTagIds.value.forEach((id) => tagIds.add(id))
    if (tagIds.size > 0) {
      params.tagIds = Array.from(tagIds)
    }
    if (dateFrom.value !== null) {
      params.dateFrom = dateFrom.value
    }
    if (dateTo.value !== null) {
      params.dateTo = dateTo.value
    }
    return params
  }

  /** 清空日期范围过滤 */
  const resetDateRange = () => {
    dateFrom.value = null
    dateTo.value = null
  }

  /** 清空面板中的多标签过滤 */
  const resetSelectedTags = () => {
    selectedTagIds.value = []
  }

  /** 从多标签过滤中移除某个标签；若该标签正好是 filteredTag，则一并退出单标签过滤模式 */
  const removeSelectedTag = (tagId: string) => {
    selectedTagIds.value = selectedTagIds.value.filter((id) => id !== tagId)
    if (filteredTag.value?.id === tagId && isFilteringMode.value) {
      isFilteringMode.value = false
      filteredTag.value = null
    }
  }

  /**
   * 加载下一页数据。
   * 通过 current > page 判断是否有新页需要加载，避免重复请求。
   * 请求成功后将数据合并进列表并推进 page 计数。
   *
   * 注意：`current <= page` 守卫是在 await 之前同步求值的，若两个调用方
   * （如 main.ts 的预热 + TheEchos.onMounted）在首次请求完成前相继进入，
   * 两者都会绕过守卫触发重复请求。为此用 `pendingFetch` 复用进行中的 Promise，
   * 保证同一 page 只发一次。
   */
  let pendingFetch: Promise<void> | null = null
  async function getEchosByPage() {
    if (current.value <= page.value) return
    if (pendingFetch) return pendingFetch

    isLoading.value = true

    pendingFetch = fetchQueryEchos(buildQueryParams())
      .then((res) => {
        if (res.code === 1) {
          total.value = res.data.total
          mergeEchoItems(echoList, echoIndexMap, res.data.items)
          page.value += 1
        }
      })
      .finally(() => {
        isLoading.value = false
        pendingFetch = null
      })

    return pendingFetch
  }

  /** 重置所有分页状态，清空列表与索引。 */
  function resetPagination() {
    current.value = 1
    page.value = 0
    echoList.value = []
    echoIndexMap.value.clear()
    total.value = 0
  }

  /** 重置分页后立即重新加载第一页。 */
  const refreshEchos = () => {
    resetPagination()
    getEchosByPage()
  }

  /** 清空列表但不重新加载（用于页面卸载等场景）。 */
  const clearEchos = () => {
    resetPagination()
  }

  /** 仅重置分页指针与列表，不触发加载（搜索场景由调用方控制加载时机）。 */
  const refreshForSearch = () => {
    current.value = 1
    page.value = 0
    echoList.value = []
    echoIndexMap.value.clear()
  }

  /** 就地更新列表中某条 Echo 的数据。 */
  const updateEcho = (echo: App.Api.Ech0.Echo) => {
    const idx = echoIndexMap.value.get(echo.id)
    if (idx !== undefined) {
      echoList.value[idx] = echo
    }
  }

  /** 乐观更新点赞数（先更新 UI，不等服务端确认）。 */
  const updateLikeCount = (echoId: string, delta: number = 1) => {
    const idx = echoIndexMap.value.get(echoId)
    if (idx !== undefined) {
      const targetEcho = echoList.value[idx]
      if (targetEcho) {
        targetEcho.fav_count = (targetEcho.fav_count || 0) + delta
        echoList.value[idx] = { ...targetEcho } // 展开赋值以触发 Vue 响应式更新
      }
    }
  }

  /** 拉取全部标签列表。 */
  const getTags = async () => {
    const res = await fetchGetTags()
    if (res.code === 1) {
      tagList.value.splice(0, tagList.value.length, ...res.data)
    }
  }

  /** 首次访问标签 UI 时懒加载；重复调用不重复请求。 */
  let tagsLoadPromise: Promise<void> | null = null
  const ensureTagsLoaded = (): Promise<void> => {
    if (tagList.value.length > 0) return Promise.resolve()
    if (tagsLoadPromise) return tagsLoadPromise
    tagsLoadPromise = getTags().finally(() => {
      tagsLoadPromise = null
    })
    return tagsLoadPromise
  }

  return {
    // 状态
    echoList,
    echoIndexMap,
    isLoading,
    total,
    pageSize,
    page,
    current,
    searchValue,
    searchingMode,
    hasMore,
    echoToUpdate,
    tagList,
    tagOptions,

    // 标签过滤
    isFilteringMode,
    filteredTag,

    // 日期范围过滤
    dateFrom,
    dateTo,
    isDateRangeActive,

    // 多标签过滤
    selectedTagIds,
    isTagSelectionActive,

    // actions
    getEchosByPage,
    resetPagination,
    refreshEchos,
    clearEchos,
    refreshForSearch,
    resetDateRange,
    resetSelectedTags,
    removeSelectedTag,
    updateEcho,
    updateLikeCount,
    getTags,
    ensureTagsLoaded,
  }
})
