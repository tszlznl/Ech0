<template>
  <div
    class="widget flex flex-col gap-2 p-4 rounded-md ring-1 ring-[var(--color-border-subtle)] ring-inset mx-auto shadow-sm hover:shadow-md"
  >
    <!-- Type -->
    <div class="inbox-header flex items-center justify-between mb-1">
      <div class="flex justify-start gap-2 items-center h-auto font-bold text-2xl">
        <Notification class="w-4 h-4" />
      </div>
      <div
        class="uppercase text-[var(--color-text-muted)] text-xs font-mono bg-[var(--color-bg-muted)] px-2 py-0.5 rounded-full"
      >
        {{ props.inbox.type }}
      </div>
    </div>

    <!-- Content -->
    <div>
      <p class="text-[var(--color-text-secondary)] text-sm whitespace-pre-wrap">
        {{ props.inbox.content }}
      </p>
    </div>

    <!-- Time & Source -->
    <div
      class="inbox-footer flex justify-between items-center mt-2 pt-2 border-t border-dashed border-[var(--color-border-strong)]"
    >
      <div class="text-[var(--color-text-muted)] text-xs font-mono">
        {{ formatDate(props.inbox.created_at) }}
      </div>
      <div class="text-xs">
        <span class="text-[var(--color-text-muted)] uppercase">{{ props.inbox.source }}</span>
      </div>
    </div>
  </div>
</template>
<script lang="ts" setup>
import Notification from '../icons/ notification.vue'
import { formatDate } from '@/utils/other'
const props = defineProps<{
  inbox: App.Api.Inbox.Inbox
}>()
</script>
<style scoped>
.inbox-header,
.inbox-footer {
  font-family: var(--font-family-display);
}
</style>
