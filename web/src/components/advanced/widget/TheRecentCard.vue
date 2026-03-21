<template>
  <div v-if="AgentSetting.enable" class="px-9 md:px-11">
    <div class="widget bg-transparent! w-full max-w-[19rem] mx-auto rounded-md p-4">
      <div class="recent-head mb-2">
        <div class="recent-icon-chip">
          <RecentIcon class="w-8 h-8" />
        </div>
        <div class="recent-title-wrap">
          <div class="recent-title">Recent</div>
          <div class="recent-title-accent">AI</div>
        </div>
      </div>

      <div class="recent-body">
        <div class="recent-card">
          <div v-if="!loading" class="recent-content">
            <TheMdPreview :content="recent" />
          </div>
          <div v-else>
            <div class="recent-loading">{{ t('recentCard.generating') }}</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import { fetchGetRecent } from '@/service/api'
import { onMounted, ref } from 'vue'
import RecentIcon from '@/components/icons/recent.vue'
import { TheMdPreview } from '@/components/advanced/md'
import { useSettingStore } from '@/stores'
import { storeToRefs } from 'pinia'
import { useI18n } from 'vue-i18n'

const settingStore = useSettingStore()
const { AgentSetting } = storeToRefs(settingStore)
const { t } = useI18n()

const recent = ref<string>(String(t('recentCard.mysteriousRecent')))
const loading = ref<boolean>(true)

onMounted(() => {
  if (AgentSetting.value.enable) {
    fetchGetRecent()
      .then((res) => {
        if (res.code === 1) {
          recent.value = res.data
        }
      })
      .finally(() => {
        loading.value = false
      })
  }
})
</script>
<style scoped>
.recent-head {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 12px;
}

.recent-icon-chip {
  width: 64px;
  height: 64px;
  border-radius: 9999px;
  color: var(--color-text-muted);
  display: flex;
  align-items: center;
  justify-content: center;
}

.recent-title-wrap {
  line-height: 0.9;
  text-align: right;
}

.recent-title {
  font-family: Georgia, 'Times New Roman', serif;
  font-size: 26px;
  font-weight: 600;
  color: var(--color-text-primary);
}

.recent-title-accent {
  font-family: 'Comic Sans MS', cursive;
  color: var(--color-accent);
  font-size: 18px;
  margin-top: -2px;
}

.recent-body {
  width: 100%;
}

.recent-card {
  position: relative;
  width: 100%;
  border: 1px solid var(--color-border-subtle);
  background-color: color-mix(in srgb, var(--color-bg-surface) 78%, transparent);
  box-shadow: 0 8px 18px rgba(20, 20, 20, 0.04);
  padding: 14px 12px 12px;
  transform: rotate(-1.1deg);
  transform-origin: top center;
  transition: transform 220ms ease;
}

.recent-card:hover {
  transform: rotate(-0.4deg);
}

.recent-card::before {
  content: '';
  position: absolute;
  left: 50%;
  top: -7px;
  transform: translateX(-50%) rotate(-1.5deg);
  width: 42px;
  height: 12px;
  border-radius: 2px;
  background: color-mix(in srgb, var(--color-bg-canvas) 84%, #d7d2bf 16%);
  box-shadow:
    0 1px 0 rgba(255, 255, 255, 0.3) inset,
    0 1px 2px rgba(0, 0, 0, 0.08);
  opacity: 0.95;
}

.recent-content {
  color: var(--color-text-secondary);
  font-size: 13px;
  line-height: 1.65;
  white-space: normal;
  word-break: break-word;
}

.recent-loading {
  color: var(--color-text-secondary);
  font-size: 13px;
}

:deep(.echo-markdown p) {
  color: var(--color-text-secondary) !important;
  font-size: 13px !important;
  line-height: 1.65 !important;
}
</style>
