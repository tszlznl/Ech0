<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="sources">
    <button
      v-for="src in visibleSources"
      :key="src.echo_id"
      class="sources__link"
      @click="emit('open', src.echo_id)"
    >
      <span class="sources__mark">↗</span>{{ formatSource(src) }}
    </button>
    <button v-if="sources.length > LIMIT" class="sources__toggle" @click="showAll = !showAll">
      {{
        showAll ? t('chatPanel.sourcesLess') : t('chatPanel.sourcesMore', { count: hiddenCount })
      }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

const props = defineProps<{
  sources: App.Api.Chat.ChatSource[]
}>()

const emit = defineEmits<{
  (e: 'open', echoId: string): void
}>()

const { t } = useI18n()

const LIMIT = 3
const showAll = ref<boolean>(false)

const visibleSources = computed(() =>
  showAll.value ? props.sources : props.sources.slice(0, LIMIT),
)
const hiddenCount = computed(() => props.sources.length - LIMIT)

const formatSource = (src: App.Api.Chat.ChatSource): string => {
  const day = new Date(src.echo_created * 1000).toISOString().slice(0, 10)
  const text = src.content.length > 40 ? src.content.slice(0, 40) + '…' : src.content
  return ` ${day} · ${text}`
}
</script>

<style scoped>
.sources {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
  margin-top: 0.35rem;
}

.sources__link {
  display: inline-flex;
  align-items: baseline;
  max-width: 32rem;
  border: none;
  background: transparent;
  padding: 0;
  font-size: 0.72rem;
  line-height: 1.5;
  color: var(--color-text-muted);
  text-align: left;
  cursor: pointer;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  transition: color 0.18s ease;
}

.sources__link:hover {
  color: var(--color-accent);
}

.sources__mark {
  color: var(--color-accent);
  opacity: 0.8;
}

.sources__toggle {
  align-self: flex-start;
  margin-top: 0.1rem;
  border: none;
  background: transparent;
  padding: 0;
  font-size: 0.72rem;
  line-height: 1.5;
  color: var(--color-text-secondary);
  cursor: pointer;
  transition: color 0.18s ease;
}

.sources__toggle:hover {
  color: var(--color-accent);
}
</style>
