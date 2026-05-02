<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="px-3 pb-4 py-2 mt-4 sm:mt-6 mb-10 mx-auto flex justify-center items-center">
    <div class="w-full sm:max-w-lg mx-auto">
      <div v-if="echo" class="w-full sm:mt-1 mx-auto">
        <TheEchoDetail :echo="echo" @update-like-count="handleUpdateLikeCount" />
        <TheEchoInteractions />
      </div>
      <div v-else class="w-full sm:mt-1 text-[var(--color-text-muted)]">
        <p class="text-center">{{ t('echoPage.loadingDetail') }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ref } from 'vue'
import TheEchoDetail from '@/components/advanced/echo/cards/TheEchoDetail.vue'
import TheEchoInteractions from '@/components/advanced/echo/cards/TheEchoInteractions.vue'
import { useEchoStore } from '@/stores'
import { useI18n } from 'vue-i18n'

const route = useRoute()
const { t } = useI18n()
const echoId = route.params.echoId as string

const echoStore = useEchoStore()
const isLoading = ref(true)
const echo = ref<App.Api.Ech0.Echo | null>(null)

// 从 echoIndexMap 获取对应的 EchoList索引
const getEchoFromStore = (): App.Api.Ech0.Echo | null => {
  const idx = echoStore.echoIndexMap.get(echoId)
  if (idx !== undefined) {
    return echoStore.echoList[idx] ?? null
  }
  return null
}

// 刷新点赞数据
const handleUpdateLikeCount = () => {
  if (echo.value) {
    // 更新 Echo 的点赞数量
    echo.value.fav_count += 1
  }
}

onMounted(async () => {
  // 先尝试从 store 获取
  echo.value = getEchoFromStore()

  // 如果 store 里没有，复用 beforeEnter 守卫已发起的请求；
  // 没走守卫的边缘场景下，prefetchEcho 会直接发起请求。
  if (!echo.value) {
    echo.value = await echoStore.prefetchEcho(echoId)
  }
  isLoading.value = false
})
</script>
