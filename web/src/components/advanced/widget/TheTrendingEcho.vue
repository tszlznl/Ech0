<template>
  <div class="trending-echo">
    <div class="trending-head">
      <div class="trending-chip">TOP 5</div>
      <div class="trending-title-wrap">
        <div class="trending-title">{{ t('dashboard.hotTitle') }}</div>
        <div class="trending-title-accent">{{ t('dashboard.hotAccent') }}</div>
      </div>
    </div>
    <div v-if="hotEchos.length" class="hot-list">
      <div v-for="(echo, idx) in hotEchos" :key="echo.id" class="hot-item">
        <span class="hot-rank">#{{ idx + 1 }}</span>
        <p class="hot-content">{{ echoSummary(echo) }}</p>
        <span class="hot-fav">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="11"
            height="11"
            viewBox="0 0 24 24"
            fill="currentColor"
            stroke="none"
          >
            <path
              d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z"
            />
          </svg>
          {{ echo.fav_count }}
        </span>
        <a class="hot-jump" :href="`/echo/${echo.id}`" target="_blank">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="13"
            height="13"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6" />
            <polyline points="15 3 21 3 21 9" />
            <line x1="10" x2="21" y1="14" y2="3" />
          </svg>
        </a>
      </div>
    </div>
    <div v-else class="hot-empty">{{ t('dashboard.noHotEchos') }}</div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { fetchGetHotEchos } from '@/service/api'

const { t } = useI18n()
const hotEchos = ref<App.Api.Ech0.Echo[]>([])

const stripHtml = (str: string) => str.replace(/<[^>]*>/g, '').trim()

const echoSummary = (echo: App.Api.Ech0.Echo): string => {
  const text = stripHtml(echo.content || '')
  if (text) return text.slice(0, 80)

  if (echo.extension) {
    const typeMap: Record<string, string> = {
      MUSIC: '🎵 Music',
      VIDEO: '🎬 Video',
      GITHUBPROJ: '📦 GitHub',
      WEBSITE: '🔗 Website',
    }
    return typeMap[echo.extension.type] || echo.extension.type
  }

  const fileCount = echo.echo_files?.length ?? 0
  if (fileCount > 0) return `🖼 ${fileCount} ${fileCount === 1 ? 'image' : 'images'}`

  return '--'
}

onMounted(async () => {
  try {
    const res = await fetchGetHotEchos(5)
    if (res.code === 1 && Array.isArray(res.data)) {
      hotEchos.value = res.data
    }
  } catch {}
})
</script>

<style scoped>
.trending-echo {
  min-width: 0;
}

.trending-head {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 0.75rem;
  margin-bottom: 0.65rem;
}

.trending-chip {
  border: 1px solid var(--color-border-subtle);
  color: var(--color-text-muted);
  font-size: 0.66rem;
  letter-spacing: 0.15em;
  padding: 0.08rem 0.45rem;
  font-family: var(--font-family-mono);
  transform: rotate(-1.8deg);
}

.trending-title-wrap {
  line-height: 0.9;
  text-align: right;
}

.trending-title {
  font-family: Georgia, 'Times New Roman', var(--font-family-display);
  font-size: 1.3rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.trending-title-accent {
  font-family: var(--font-family-handwritten);
  color: var(--color-accent);
  font-size: 0.95rem;
  margin-top: -2px;
}

.hot-list {
  display: flex;
  flex-direction: column;
}

.hot-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.45rem 0;
  border-bottom: 1px dashed var(--color-border-subtle);
}

.hot-item:last-child {
  border-bottom: none;
}

.hot-rank {
  flex-shrink: 0;
  width: 1.6rem;
  font-size: 0.7rem;
  font-family: var(--font-family-mono);
  color: var(--color-text-muted);
  letter-spacing: -0.02em;
}

.hot-content {
  margin: 0;
  font-size: 0.82rem;
  color: var(--color-text-secondary);
  line-height: 1.45;
  font-family: var(--font-family-display);
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
}

.hot-fav {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  gap: 0.15rem;
  font-size: 0.68rem;
  color: var(--color-text-muted);
  font-family: var(--font-family-mono);
  white-space: nowrap;
}

.hot-fav svg {
  opacity: 0.45;
}

.hot-jump {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-muted);
  opacity: 0;
  transition:
    opacity 0.15s ease,
    color 0.15s ease;
}

.hot-item:hover .hot-jump {
  opacity: 1;
}

.hot-jump:hover {
  color: var(--color-accent);
}

.hot-empty {
  color: var(--color-text-muted);
  font-size: 0.82rem;
  font-family: var(--font-family-display);
  padding: 0.5rem 0;
}
</style>
