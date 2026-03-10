<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { fetchGetEchosByPage, fetchGetConnectList, fetchUnreadInbox } from '@/service/api'
import { useSettingStore } from '@/stores'

type StatCard = {
  key: string
  label: string
  value: string
}

const settingStore = useSettingStore()

const loading = ref(true)
const echoTotal = ref<number | null>(null)
const unreadInboxCount = ref<number | null>(null)
const connectCount = ref<number | null>(null)

const dashboardStats = computed<StatCard[]>(() => {
  return [
    {
      key: 'echos',
      label: 'Echo 总数',
      value: echoTotal.value === null ? '--' : String(echoTotal.value),
    },
    {
      key: 'inbox',
      label: '未读收件箱',
      value: unreadInboxCount.value === null ? '--' : String(unreadInboxCount.value),
    },
    {
      key: 'connect',
      label: '已连接节点',
      value: connectCount.value === null ? '--' : String(connectCount.value),
    },
    {
      key: 'version',
      label: '当前版本',
      value: settingStore.hello?.version || '--',
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
    <section class="stats-grid">
      <article v-for="item in dashboardStats" :key="item.key" class="stat-card">
        <p class="stat-label">{{ item.label }}</p>
        <p class="stat-value" :class="{ 'is-loading': loading && item.value === '--' }">{{ item.value }}</p>
      </article>
    </section>
  </div>
</template>

<style scoped>
.dashboard-page {
  width: 100%;
  padding: 0.4rem 0.25rem 0.9rem;
}

.stats-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 0.85rem;
}

.stat-card {
  border: 1px solid var(--color-border-subtle);
  border-radius: var(--radius-lg);
  background: var(--color-bg-surface);
  padding: 1rem 1.05rem;
  transition:
    border-color 0.2s ease,
    box-shadow 0.2s ease,
    background-color 0.2s ease;
}

.stat-card:hover {
  border-color: color-mix(in oklab, var(--color-accent) 28%, var(--color-border-subtle));
  box-shadow: 0 3px 10px color-mix(in oklab, var(--color-accent) 6%, transparent);
  background: color-mix(in oklab, var(--color-bg-surface) 92%, var(--color-accent-soft));
}

.stat-label {
  font-size: 0.88rem;
  color: var(--color-text-muted);
}

.stat-value {
  margin-top: 0.25rem;
  font-size: 1.5rem;
  line-height: 1.2;
  color: var(--color-accent);
  font-weight: 800;
  font-family: var(--font-family-display);
}

.stat-value.is-loading {
  opacity: 0.7;
}

@media (min-width: 768px) {
  .stats-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
