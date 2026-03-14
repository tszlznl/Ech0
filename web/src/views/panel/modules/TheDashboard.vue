<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { fetchGetEchosByPage, fetchGetConnectList, fetchUnreadInbox } from '@/service/api'
import { useSettingStore } from '@/stores'
import PanelCard from '@/layout/PanelCard.vue'

type StatCard = {
  key: string
  serial: string
  label: string
  value: string
  note: string
}

type StatusCard = {
  key: string
  title: string
  value: string
}

const settingStore = useSettingStore()
const { t, locale } = useI18n()

const loading = ref(true)
const echoTotal = ref<number | null>(null)
const unreadInboxCount = ref<number | null>(null)
const connectCount = ref<number | null>(null)

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
      key: 'inbox',
      serial: 'NO.02',
      label: String(t('dashboard.unreadInbox')),
      value: formatMetric(unreadInboxCount.value),
      note: '',
    },
    {
      key: 'connect',
      serial: 'NO.03',
      label: String(t('dashboard.connectedNodes')),
      value: formatMetric(connectCount.value),
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

const dashboardStatus = computed<StatusCard[]>(() => {
  const unread = unreadInboxCount.value ?? 0
  const connect = connectCount.value ?? 0

  return [
    {
      key: 'inbox-status',
      title: String(t('dashboard.inboxStatus')),
      value:
        unread > 0
          ? String(t('dashboard.pendingCount', { count: unread }))
          : String(t('dashboard.cleared')),
    },
    {
      key: 'connect-status',
      title: String(t('dashboard.connectionStatus')),
      value:
        connect > 0
          ? String(t('dashboard.nodesOnline', { count: connect }))
          : String(t('dashboard.noNodes')),
    },
  ]
})

const loadDashboardStats = async () => {
  loading.value = true
  const [echoRes, unreadRes, connectRes] = await Promise.allSettled([
    fetchGetEchosByPage({ page: 1, pageSize: 1, search: '' }),
    fetchUnreadInbox(),
    fetchGetConnectList(),
  ])

  if (echoRes.status === 'fulfilled' && echoRes.value.code === 1) {
    echoTotal.value = echoRes.value.data?.total ?? 0
  }

  if (unreadRes.status === 'fulfilled' && unreadRes.value.code === 1) {
    unreadInboxCount.value = Array.isArray(unreadRes.value.data) ? unreadRes.value.data.length : 0
  }

  if (connectRes.status === 'fulfilled' && connectRes.value.code === 1) {
    connectCount.value = Array.isArray(connectRes.value.data) ? connectRes.value.data.length : 0
  }

  loading.value = false
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
        class="stat-card"
      >
        <div class="stat-head">
          <p class="stat-serial">{{ item.serial }}</p>
          <span class="stat-head-line"></span>
        </div>
        <p class="stat-label">{{ item.label }}</p>
        <p class="stat-value" :class="{ 'is-loading': loading && item.value === '--' }">
          {{ item.value }}
        </p>
        <p class="stat-note">{{ item.note }}</p>
      </PanelCard>
    </section>

    <section class="status-grid">
      <PanelCard
        v-for="item in dashboardStatus"
        :key="item.key"
        border-style="solid"
        class="status-card"
      >
        <div class="status-head">
          <p class="status-title">{{ item.title }}</p>
          <span class="status-head-line"></span>
        </div>
        <p class="status-value">{{ item.value }}</p>
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

.status-grid {
  display: grid;
  grid-template-columns: repeat(1, minmax(0, 1fr));
  gap: 0.7rem;
}

.status-card {
  display: grid;
  gap: 0.35rem;
  padding: 0.7rem 0.9rem;
  box-shadow: none;
}

.status-head {
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.status-title {
  color: var(--color-text-muted);
  font-size: 0.78rem;
  white-space: nowrap;
}

.status-head-line {
  display: inline-block;
  width: 100%;
  height: 1px;
  background: color-mix(in oklab, var(--color-border-subtle) 85%, transparent);
}

.status-value {
  color: var(--color-text-secondary);
  font-size: 0.8rem;
  font-weight: 600;
}

@media (min-width: 768px) {
  .stats-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .status-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .dashboard-page {
    gap: 0.85rem;
  }
}
</style>
