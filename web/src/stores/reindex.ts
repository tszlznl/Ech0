// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { fetchCancelReindex, fetchReindexEmbeddings, fetchReindexStatus } from '@/service/api'

// 复用 migration 的轮询范式：起作业后每 2s 轮询状态，非进行中即停。
const POLL_INTERVAL_MS = 2000

type ReindexStatusValue = App.Api.Embedding.ReindexStatus['status']

export const useReindexStore = defineStore('reindexStore', () => {
  const status = ref<ReindexStatusValue>('idle')
  const phase = ref('')
  const error = ref('')
  const result = ref<App.Api.Embedding.ReindexResult | null>(null)
  const pollTimer = ref<number | null>(null)

  const isRunning = computed(() => status.value === 'pending' || status.value === 'running')

  function applyState(next: App.Api.Embedding.ReindexStatus | null | undefined) {
    if (!next) return
    status.value = next.status
    phase.value = next.phase ?? ''
    error.value = next.error ?? ''
    result.value = next.payload ?? null
  }

  function stopPolling() {
    if (pollTimer.value !== null) {
      window.clearInterval(pollTimer.value)
      pollTimer.value = null
    }
  }

  function startPolling() {
    if (pollTimer.value !== null) return
    pollTimer.value = window.setInterval(async () => {
      await fetchStatus()
      if (!isRunning.value) {
        stopPolling()
      }
    }, POLL_INTERVAL_MS)
  }

  async function fetchStatus() {
    const res = await fetchReindexStatus()
    if (res.code !== 1) {
      return false
    }
    applyState(res.data as App.Api.Embedding.ReindexStatus)
    return true
  }

  async function start() {
    const res = await fetchReindexEmbeddings()
    if (res.code !== 1) {
      return res
    }
    applyState(res.data as App.Api.Embedding.ReindexStatus)
    if (isRunning.value) {
      startPolling()
    }
    return res
  }

  async function cancel() {
    const res = await fetchCancelReindex()
    if (res.code === 1) {
      applyState(res.data as App.Api.Embedding.ReindexStatus)
      // 取消是协作式的：作业可能尚未落到 cancelled，需继续轮询直到后端确认终态，
      // 否则会停在 running 不再刷新。
      if (isRunning.value) {
        startPolling()
      } else {
        stopPolling()
      }
    }
    return res
  }

  // 页面挂载时拉取一次：若重建在跑（含刷新后续显），恢复轮询。
  async function init() {
    const ok = await fetchStatus()
    if (ok && isRunning.value) {
      startPolling()
    }
  }

  return {
    status,
    phase,
    error,
    result,
    isRunning,
    init,
    fetchStatus,
    start,
    cancel,
    startPolling,
    stopPolling,
  }
})
