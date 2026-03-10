<template>
  <div>
    <h2 class="text-[var(--color-text-secondary)] font-bold mb-1">收件箱模式</h2>

    <!-- 当前共有 -->
    <div class="text-[var(--color-text-muted)] my-2 inbox-mode-metrics">
      <p>
        当前消息共有 <span class="font-bold">{{ total }}</span> 条
      </p>
    </div>

    <!-- 未读消息共有 -->
    <div class="text-[var(--color-text-muted)] my-2 inbox-mode-metrics">
      <p>
        未读消息共有 <span class="font-bold">{{ unreadItems.length }}</span> 条
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { useInboxStore } from '@/stores'

const inboxStore = useInboxStore()
const { unreadItems, total } = storeToRefs(inboxStore)
</script>

<style scoped>
.inbox-mode-metrics {
  font-family: var(--font-family-display);
}
</style>
