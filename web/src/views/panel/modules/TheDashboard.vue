<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { fetchGetEchosByPage, fetchGetTodayEchos, fetchCheckUpdate } from '@/service/api'
import { useConnectStore, useSettingStore } from '@/stores'
import { theToast } from '@/utils/toast'
import PanelCard from '@/layout/PanelCard.vue'

type StatCard = {
  key: string
  serial: string
  label: string
  value: string
  note: string
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
  if (value === null) {
    return '--'
  }

  return new Intl.NumberFormat(locale.value).format(value)
}

const dashboardStats = computed<StatCard[]>(() => {
  return [
    {
      key: 'echos',
      serial: 'NO.01',
      label: String(t('dashboard.echoTotal')),
      value: formatMetric(echoTotal.value),
      note: '',
    },
    {
      key: 'connect',
      serial: 'NO.02',
      label: String(t('dashboard.connectedNodes')),
      value: formatMetric(connectCount.value),
      note: '',
    },
    {
      key: 'today-echo',
      serial: 'NO.03',
      label: String(t('dashboard.todayEchoCount')),
      value: formatMetric(todayEchoCount.value),
      note: '',
    },
    {
      key: 'version',
      serial: 'NO.04',
      label: String(t('dashboard.currentVersion')),
      value: settingStore.hello?.version || '--',
      note: '',
    },
  ]
})

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
    todayEchoCount.value = Array.isArray(todayRes.value.data) ? todayRes.value.data.length : 0
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
  if (key === 'version') {
    void handleCheckUpdate()
  }
}

onMounted(() => {
  void loadDashboardStats()
})
</script>

<template>
  <div class="dashboard-page">
    <section class="dashboard-meta">
      <span class="meta-item meta-item-strong">PANEL DASHBOARD</span>
      <span class="meta-item">DATE {{ todayText }}</span>
      <span class="meta-item">VERSION {{ settingStore.hello?.version || '--' }}</span>
    </section>

    <section class="stats-grid">
      <PanelCard
        v-for="item in dashboardStats"
        :key="item.key"
        border-style="solid"
        :class="['stat-card', item.key === 'version' ? 'stat-card--clickable' : '']"
        v-tooltip="item.key === 'version' ? t('dashboard.clickToCheckUpdate') : undefined"
        @click="handleStatCardClick(item.key)"
      >
        <div class="stat-head">
          <p class="stat-serial">{{ item.serial }}</p>
          <span class="stat-head-line"></span>
        </div>
        <p class="stat-label">{{ item.label }}</p>
        <p class="stat-value" :class="{ 'is-loading': loading && item.value === '--' }">
          {{ item.value }}
          <span
            v-if="item.key === 'version' && hasUpdate"
            class="update-dot"
            :title="t('dashboard.updateAvailable', { version: latestVersion })"
          ></span>
          <span v-if="item.key === 'version' && checkingUpdate" class="stat-checking">{{
            t('dashboard.checkingUpdate')
          }}</span>
        </p>
        <p class="stat-note">{{ item.note }}</p>
      </PanelCard>
    </section>
  </div>
</template>

<style scoped>
.dashboard-page {
  width: 100%;
  padding: 0.35rem 0.25rem 1rem;
  display: grid;
  gap: 0.75rem;
}

.dashboard-meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.35rem 0.5rem;
  border: 1px solid var(--color-border-subtle);
  border-radius: var(--radius-lg);
  padding: 0.45rem 0.6rem;
  background: color-mix(in oklab, var(--color-bg-surface) 95%, var(--color-bg-muted));
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
  grid-template-columns: repeat(1, minmax(0, 1fr));
  gap: 0.7rem;
}

.stat-card {
  padding: 0.9rem 0.95rem;
  box-shadow: none;
  transition:
    border-color 0.2s ease,
    box-shadow 0.2s ease,
    background-color 0.2s ease;
}

.stat-card:hover {
  border-color: var(--color-border-strong);
  box-shadow: 0 2px 8px color-mix(in oklab, var(--color-text-primary) 4%, transparent);
  background: color-mix(in oklab, var(--color-bg-surface) 96%, var(--color-bg-muted));
}

.stat-head {
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.stat-serial {
  font-size: 0.72rem;
  color: var(--color-text-muted);
  letter-spacing: 0.04em;
  font-family: var(--font-family-mono);
  text-transform: uppercase;
  white-space: nowrap;
}

.stat-head-line {
  display: inline-block;
  width: 100%;
  height: 1px;
  background: color-mix(in oklab, var(--color-border-subtle) 85%, transparent);
}

.stat-label {
  margin-top: 0.2rem;
  font-size: 0.82rem;
  color: var(--color-text-secondary);
  line-height: 1.35;
}

.stat-value {
  margin-top: 0.25rem;
  font-size: clamp(1.35rem, 3.1vw, 1.8rem);
  line-height: 1.2;
  color: var(--color-text-primary);
  font-weight: 700;
  font-family: var(--font-family-display);
}

.stat-note {
  margin-top: 0.35rem;
  color: var(--color-text-muted);
  font-size: 0.76rem;
  line-height: 1.3;
}

.stat-value.is-loading {
  opacity: 0.7;
}

.stat-card--clickable {
  cursor: pointer;
}

.update-dot {
  display: inline-block;
  width: 0.5rem;
  height: 0.5rem;
  border-radius: 50%;
  background: #e5a00d;
  margin-left: 0.4rem;
  vertical-align: middle;
  animation: update-dot-pulse 2s ease-in-out infinite;
}

.stat-checking {
  font-size: 0.6em;
  color: var(--color-text-muted);
  margin-left: 0.2em;
  animation: stat-checking-blink 1s steps(1, end) infinite;
}

@keyframes update-dot-pulse {
  0%,
  100% {
    opacity: 1;
  }
  50% {
    opacity: 0.4;
  }
}

@keyframes stat-checking-blink {
  0%,
  50% {
    opacity: 1;
  }
  51%,
  100% {
    opacity: 0.3;
  }
}

@media (min-width: 768px) {
  .stats-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .dashboard-page {
    gap: 0.85rem;
  }
}
</style>
