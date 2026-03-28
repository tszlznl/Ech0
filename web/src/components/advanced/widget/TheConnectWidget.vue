<template>
  <div class="px-9 md:px-11">
    <div class="widget bg-transparent! w-full max-w-[19rem] mx-auto rounded-md p-4">
      <div class="connect-head mb-2">
        <div class="connect-icon-chip">
          <Connect class="w-8 h-8" />
        </div>
        <div class="connect-title-wrap">
          <div class="connect-title">Connect</div>
          <div class="connect-title-accent">Widget</div>
        </div>
      </div>
      <div v-if="!loading">
        <div v-if="!connectsInfo.length" class="text-[var(--color-text-muted)] text-sm mb-2">
          {{ t('connectWidget.noConnections') }}
        </div>
        <div v-else class="flex flex-wrap gap-3">
          <div
            v-for="(connect, index) in connectsInfo"
            :key="index"
            class="relative flex flex-col items-center justify-center w-8 h-8 min-w-[2rem] min-h-[2rem] flex-none border-2 border-[var(--color-border-subtle)] shadow-sm rounded-full hover:shadow-md transition duration-200 ease-in-out group"
          >
            <a :href="connect.server_url" target="_blank" class="block w-full h-full">
              <img
                :src="connect.logo"
                alt="avatar"
                loading="lazy"
                decoding="async"
                class="w-full h-full rounded-full object-cover"
              />
              <span
                class="absolute top-0 right-0 w-2.5 h-2.5 border-2 border-[var(--color-bg-surface)] rounded-full"
                :style="{
                  transform: 'translate(35%, -35%)',
                  backgroundColor: getColor(connect.today_echos || 0),
                }"
              ></span>
            </a>
            <div
              class="absolute z-10 left-1/2 -translate-x-1/2 top-10 min-w-max bg-gray-800 text-white text-xs rounded px-3 py-2 opacity-0 group-hover:opacity-100 pointer-events-none transition-opacity duration-200 shadow-lg"
            >
              <div class="font-bold mb-1">{{ connect.server_name }}</div>
              <div>{{ t('connectWidget.tooltipOwner') }}: {{ connect.sys_username || '-' }}</div>
              <div>{{ t('connectWidget.tooltipTotal') }}: {{ connect.total_echos ?? 0 }}</div>
              <div>{{ t('connectWidget.tooltipToday') }}: {{ connect.today_echos ?? 0 }}</div>
              <div>{{ t('connectWidget.tooltipVersion') }}: {{ connect.version || '-' }}</div>
            </div>
          </div>
        </div>
      </div>
      <div v-else>
        <div class="text-[var(--color-text-secondary)] text-sm mb-2">
          {{ t('connectWidget.loading') }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Connect from '@/components/icons/connect.vue'
import { useConnectStore } from '@/stores'
import { storeToRefs } from 'pinia'
import { onMounted } from 'vue'
import { useI18n } from 'vue-i18n'

const connectStore = useConnectStore()
const { t } = useI18n()
const { getConnectInfo } = connectStore
const { loading, connectsInfo } = storeToRefs(connectStore)

const getColor = (count: number): string => {
  if (count >= 4) return 'var(--color-accent)'
  if (count >= 3) return 'var(--color-accent)'
  if (count >= 2) return 'var(--color-accent)'
  if (count >= 1) return 'var(--color-accent-soft)'
  return '#c4c3c1'
}

onMounted(() => {
  getConnectInfo()
})
</script>

<style scoped>
.connect-head {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 12px;
}

.connect-icon-chip {
  width: 64px;
  height: 64px;
  border-radius: 9999px;
  color: var(--color-text-muted);
  display: flex;
  align-items: center;
  justify-content: center;
}

.connect-title-wrap {
  line-height: 0.9;
  text-align: right;
}

.connect-title {
  font-family: Georgia, 'Times New Roman', serif;
  font-size: 26px;
  font-weight: 600;
  color: var(--color-text-primary);
}

.connect-title-accent {
  font-family: var(--font-family-handwritten);
  color: var(--color-accent);
  font-size: 18px;
  margin-top: -2px;
}
</style>
