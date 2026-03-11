<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
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

const loading = ref(true)
const echoTotal = ref<number | null>(null)
const unreadInboxCount = ref<number | null>(null)
const connectCount = ref<number | null>(null)

const formatMetric = (value: number | null) => {
  if (value === null) {
    return '--'
  }

  return new Intl.NumberFormat('zh-CN').format(value)
}

const dashboardStats = computed<StatCard[]>(() => {
  return [
    {
      key: 'echos',
      serial: 'NO.01',
      label: 'Echo 总数',
      value: formatMetric(echoTotal.value),
      note: '累计灵感条目',
    },
    {
      key: 'inbox',
      serial: 'NO.02',
      label: '未读收件箱',
      value: formatMetric(unreadInboxCount.value),
      note: '等待你处理的消息',
    },
    {
      key: 'connect',
      serial: 'NO.03',
      label: '已连接节点',
      value: formatMetric(connectCount.value),
      note: '当前在线连接能力',
    },
    {
      key: 'version',
      serial: 'NO.04',
      label: '当前版本',
      value: settingStore.hello?.version || '--',
      note: '客户端运行版本',
    },
  ]
})

const todayText = computed(() => {
  return new Intl.DateTimeFormat('zh-CN', {
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
      title: '收件箱状态',
      value: unread > 0 ? `待处理 ${unread}` : '已清空',
    },
    {
      key: 'connect-status',
      title: '连接状态',
      value: connect > 0 ? `${connect} 个节点在线` : '暂无可用节点',
    },
  ]
})

const dashboardInsights = computed(() => {
  if (loading.value) {
    return ['同步中：正在加载数据摘要', '同步中：状态短句即将更新']
  }

  const insights: string[] = []
  const unread = unreadInboxCount.value ?? 0
  const connect = connectCount.value ?? 0
  const echos = echoTotal.value ?? 0

  insights.push(`收件箱：${unread > 0 ? `待处理 ${unread}` : '无待处理消息'}`)
  insights.push(`节点：${connect > 0 ? `${connect} 个连接可用` : '未检测到连接节点'}`)
  insights.push(`记录：累计 Echo ${echos} 条`)

  return insights
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
      <PanelCard v-for="item in dashboardStats" :key="item.key" border-style="solid" class="stat-card">
        <div class="stat-head">
          <p class="stat-serial">{{ item.serial }}</p>
          <span class="stat-head-line"></span>
        </div>
        <p class="stat-label">{{ item.label }}</p>
        <p class="stat-value" :class="{ 'is-loading': loading && item.value === '--' }">{{ item.value }}</p>
        <p class="stat-note">{{ item.note }}</p>
      </PanelCard>
    </section>

    <section class="status-grid">
      <PanelCard v-for="item in dashboardStatus" :key="item.key" border-style="solid" class="status-card">
        <div class="status-head">
          <p class="status-title">{{ item.title }}</p>
          <span class="status-head-line"></span>
        </div>
        <p class="status-value">{{ item.value }}</p>
      </PanelCard>
    </section>

    <PanelCard border-style="solid" class="dashboard-insights">
      <div class="insight-head">
        <span class="insight-tag">STATUS NOTES</span>
        <span class="insight-head-line"></span>
      </div>
      <ul class="insight-list">
        <li v-for="(item, index) in dashboardInsights" :key="index" class="insight-item">
          {{ item }}
        </li>
      </ul>
    </PanelCard>
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

.dashboard-insights {
  background: var(--color-bg-surface);
  padding: 0.8rem 0.95rem;
  box-shadow: none;
}

.insight-head {
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.insight-tag {
  font-size: 0.72rem;
  letter-spacing: 0.04em;
  color: var(--color-text-secondary);
  font-family: var(--font-family-mono);
  white-space: nowrap;
}

.insight-head-line {
  display: inline-block;
  width: 100%;
  height: 1px;
  background: color-mix(in oklab, var(--color-border-subtle) 85%, transparent);
}

.insight-list {
  margin-top: 0.55rem;
  display: grid;
  gap: 0.35rem;
}

.insight-item {
  border-left: 1px solid var(--color-border-subtle);
  padding-left: 0.5rem;
  color: var(--color-text-secondary);
  font-size: 0.82rem;
  line-height: 1.45;
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
