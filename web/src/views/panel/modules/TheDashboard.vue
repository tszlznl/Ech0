<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { fetchGetEchosByPage, fetchGetTodayEchos, fetchCheckUpdate } from '@/service/api'
import { useConnectStore, useSettingStore } from '@/stores'
import { theToast } from '@/utils/toast'
import { TheActivityLog, TheTrendingEcho } from '@/components/advanced/widget'
import Box from '@/components/icons/box.vue'
import ConnectedIcon from '@/components/icons/connected-icon.vue'
import DateIcon from '@/components/icons/date-icon.vue'
import TotalIcon from '@/components/icons/total.vue'

type StatCard = {
  key: string
  label: string
  value: string
  icon: string
}

const settingStore = useSettingStore()
const connectStore = useConnectStore()
const { t, locale } = useI18n()

const loading = ref(true)
const echoTotal = ref<number | null>(null)
const todayEchoCount = ref<number | null>(null)
const connectCount = ref<number | null>(null)
const hasUpdate = ref(false)
const latestVersion = ref('')
const checkingUpdate = ref(false)

const formatMetric = (value: number | null) => {
  if (value === null) return '--'
  return new Intl.NumberFormat(locale.value).format(value)
}

const dashboardStats = computed<StatCard[]>(() => [
  {
    key: 'echos',
    label: String(t('dashboard.echoTotal')),
    value: formatMetric(echoTotal.value),
    icon: 'echos',
  },
  {
    key: 'today-echo',
    label: String(t('dashboard.todayEchoCount')),
    value: formatMetric(todayEchoCount.value),
    icon: 'today',
  },
  {
    key: 'connect',
    label: String(t('dashboard.connectedNodes')),
    value: formatMetric(connectCount.value),
    icon: 'connect',
  },
  {
    key: 'version',
    label: String(t('dashboard.currentVersion')),
    value: settingStore.hello?.version || '--',
    icon: 'version',
  },
])

const todayText = computed(() => {
  return new Intl.DateTimeFormat(locale.value, {
    month: '2-digit',
    day: '2-digit',
    weekday: 'long',
  }).format(new Date())
})

const loadDashboardStats = async () => {
  loading.value = true
  const [echoRes, , todayRes] = await Promise.allSettled([
    fetchGetEchosByPage({ page: 1, pageSize: 1, search: '' }),
    connectStore.getConnect(),
    fetchGetTodayEchos(),
  ])

  if (echoRes.status === 'fulfilled' && echoRes.value.code === 1) {
    echoTotal.value = echoRes.value.data?.total ?? 0
  }

  if (todayRes.status === 'fulfilled' && todayRes.value.code === 1) {
    const data = todayRes.value.data
    todayEchoCount.value = Array.isArray(data) ? data.length : 0
  }

  connectCount.value = connectStore.connects.length
  loading.value = false
}

const handleCheckUpdate = async () => {
  if (checkingUpdate.value) return
  checkingUpdate.value = true
  try {
    const res = await fetchCheckUpdate()
    if (res.code === 1 && res.data) {
      hasUpdate.value = res.data.has_update
      latestVersion.value = res.data.latest_version
      if (res.data.has_update) {
        theToast.info(String(t('dashboard.updateAvailable', { version: res.data.latest_version })))
      } else {
        theToast.info(String(t('dashboard.alreadyLatest')))
      }
    } else {
      theToast.error(String(t('dashboard.checkUpdateFailed')))
    }
  } catch {
    theToast.error(String(t('dashboard.checkUpdateFailed')))
  } finally {
    checkingUpdate.value = false
  }
}

const handleStatCardClick = (key: string) => {
  if (key === 'version') void handleCheckUpdate()
}

onMounted(() => void loadDashboardStats())
</script>

<template>
  <div class="dashboard-page">
    <!-- Meta bar -->
    <section class="dashboard-meta">
      <span class="meta-item meta-item-strong">PANEL DASHBOARD</span>
      <span class="meta-item">DATE {{ todayText }}</span>
      <span class="meta-item">VERSION {{ settingStore.hello?.version || '--' }}</span>
    </section>

    <!-- Stat cards -->
    <section class="stats-grid">
      <div
        v-for="item in dashboardStats"
        :key="item.key"
        :class="['stat-card', item.key === 'version' ? 'stat-card--clickable' : '']"
        v-tooltip="item.key === 'version' ? t('dashboard.clickToCheckUpdate') : undefined"
        @click="handleStatCardClick(item.key)"
      >
        <div class="stat-icon-wrap">
          <TotalIcon v-if="item.icon === 'echos'" class="stat-card-icon stat-card-icon--fill" />
          <DateIcon v-else-if="item.icon === 'today'" class="stat-card-icon stat-card-icon--fill" />
          <ConnectedIcon
            v-else-if="item.icon === 'connect'"
            class="stat-card-icon stat-card-icon--stroke"
          />
          <Box v-else-if="item.icon === 'version'" class="stat-card-icon stat-card-icon--stroke" />
        </div>
        <div class="stat-body">
          <p class="stat-value" :class="{ 'is-loading': loading && item.value === '--' }">
            {{ item.value }}
            <span
              v-if="item.key === 'version' && hasUpdate"
              class="update-dot"
              :title="t('dashboard.updateAvailable', { version: latestVersion })"
            />
            <svg
              v-if="item.key === 'version' && checkingUpdate"
              class="stat-spinner"
              xmlns="http://www.w3.org/2000/svg"
              width="14"
              height="14"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2.5"
              stroke-linecap="round"
            >
              <path d="M21 12a9 9 0 1 1-6.219-8.56" />
            </svg>
          </p>
          <p class="stat-label">{{ item.label }}</p>
        </div>
      </div>
    </section>

    <!-- Panels -->
    <section class="dashboard-panels">
      <TheActivityLog />
      <TheTrendingEcho />
    </section>
  </div>
</template>

<style scoped>
.dashboard-page {
  width: 100%;
  max-width: 100%;
  overflow: hidden;
  padding: 0.35rem 0 1rem;
  display: grid;
  gap: 0.85rem;
}

.dashboard-meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.35rem 0.5rem;
  border-bottom: 1px dashed var(--color-border-subtle);
  padding: 0.35rem 0.2rem 0.55rem;
}

.meta-item {
  font-size: 0.73rem;
  letter-spacing: 0.04em;
  color: var(--color-text-muted);
  font-family: var(--font-family-mono);
}

.meta-item-strong {
  color: var(--color-text-secondary);
  font-weight: 700;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 0.1rem 0.6rem;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 0.7rem;
  padding: 0.65rem 0.4rem;
  border-bottom: 1px dashed var(--color-border-subtle);
  transition: opacity 0.2s ease;
}

.stat-card:hover {
  opacity: 0.75;
}

.stat-card--clickable {
  cursor: pointer;
}

.stat-icon-wrap {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 2.2rem;
  height: 2.2rem;
  border-radius: var(--radius-md);
  border: 1px dashed var(--color-border-subtle);
  color: var(--color-text-muted);
}

.stat-card-icon {
  display: block;
  width: 18px;
  height: 18px;
}

.stat-card-icon--fill :deep(path) {
  fill: currentColor;
}

.stat-card-icon--stroke :deep(path) {
  stroke: currentColor;
}

.stat-card-icon--stroke :deep(g) {
  stroke: currentColor;
}

.stat-body {
  min-width: 0;
}

.stat-value {
  margin: 0;
  font-size: clamp(1.15rem, 2.4vw, 1.4rem);
  line-height: 1.2;
  color: var(--color-text-primary);
  font-weight: 600;
  font-family: var(--font-family-display);
  letter-spacing: -0.01em;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.stat-value.is-loading {
  opacity: 0.5;
}

.stat-label {
  margin: 0.05rem 0 0;
  font-size: 0.72rem;
  color: var(--color-text-muted);
  line-height: 1.3;
  font-family: var(--font-family-display);
  letter-spacing: 0.01em;
}

.update-dot {
  display: inline-block;
  width: 0.42rem;
  height: 0.42rem;
  border-radius: 50%;
  background: var(--color-accent);
  margin-left: 0.3rem;
  vertical-align: middle;
  animation: dot-pulse 2s ease-in-out infinite;
}

.stat-spinner {
  display: inline-block;
  margin-left: 0.25em;
  vertical-align: middle;
  color: var(--color-text-muted);
  animation: spin 0.8s linear infinite;
}

.dashboard-panels {
  display: grid;
  grid-template-columns: 1fr;
  gap: 1rem;
}

@keyframes dot-pulse {
  0%,
  100% {
    opacity: 1;
  }
  50% {
    opacity: 0.35;
  }
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

@media (min-width: 768px) {
  .stats-grid {
    grid-template-columns: repeat(4, 1fr);
  }

  .dashboard-panels {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 400px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
}
</style>
