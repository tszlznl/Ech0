<template>
  <div class="mx-auto px-2 sm:px-5 my-4">
    <div v-for="item in items" :key="item.id" class="mb-3">
      <TheInboxCard :inbox="item" />
    </div>

    <!-- 加载更多 -->
    <div v-if="hasMore && !loading" class="my-4 ml-1 flex items-center justify-start">
      <BaseButton
        @click="loadMore"
        class="rounded-full bg-[var(--btn-bg-color)] !active:bg-[var(--btn-hover-bg-color)] mr-2"
      >
        <span
          class="text-[var(--btn-text-color)] text-md inbox-load-more-text text-center px-2 py-1"
          >加载更多</span
        >
      </BaseButton>
    </div>
    <!-- 没有更多 -->
    <div v-if="!hasMore && !loading" class="mx-auto my-5 text-center">
      <p class="text-xl text-[var(--color-text-muted)]">
        {{ t('inbox.empty') }}
      </p>
    </div>
    <!-- 加载中 -->
    <TheLoadingIndicator v-if="loading" class="mx-auto my-5" size="lg" :label="t('inbox.loading')" />
  </div>
</template>
<script lang="ts" setup>
import { onMounted, onUnmounted } from 'vue'
import { storeToRefs } from 'pinia'
import { useInboxStore } from '@/stores'
import TheInboxCard from '@/components/advanced/TheInboxCard.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import TheLoadingIndicator from '@/components/common/TheLoadingIndicator.vue'
import { useI18n } from 'vue-i18n'

const inboxStore = useInboxStore()
const { t } = useI18n()
const { items, hasMore, loading } = storeToRefs(inboxStore)
const { loadMore, markAsRead } = inboxStore

let timer: ReturnType<typeof setInterval>
const markingIds = new Set<string>()

onMounted(async () => {
  // 用户停留超过 1 秒则更新消息为已读
  timer = setInterval(() => {
    if (items.value.length > 0) {
      items.value.forEach((item) => {
        if (!item.read && !markingIds.has(item.id)) {
          markingIds.add(item.id)
          markAsRead(item.id).finally(() => {
            markingIds.delete(item.id)
          })
        }
      })
    }
  }, 1500)
})

onUnmounted(() => {
  clearInterval(timer)
  markingIds.clear()
})
</script>
<style scoped>
.inbox-load-more-text {
  font-family: var(--font-family-display);
}
</style>
