<template>
  <div>
    <h2 class="text-[var(--color-text-secondary)] font-bold mb-1">{{ t('editor.inboxMode') }}</h2>

    <!-- 当前共有 -->
    <div class="text-[var(--color-text-muted)] my-2 inbox-mode-metrics">
      <p>
        {{ t('editor.inboxTotal', { total }) }}
      </p>
    </div>

    <!-- 未读消息共有 -->
    <div class="text-[var(--color-text-muted)] my-2 inbox-mode-metrics">
      <p>
        {{ t('editor.inboxUnread', { total: unreadItems.length }) }}
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { useInboxStore } from '@/stores'
import { useI18n } from 'vue-i18n'

const inboxStore = useInboxStore()
const { unreadItems, total } = storeToRefs(inboxStore)
const { t } = useI18n()
</script>

<style scoped>
.inbox-mode-metrics {
  font-family: var(--font-family-display);
}
</style>
