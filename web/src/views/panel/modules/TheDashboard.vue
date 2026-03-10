<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { fetchGetEchosByPage, fetchGetConnectList, fetchUnreadInbox } from '@/service/api'
import { useUserStore, useSettingStore } from '@/stores'

type StatCard = {
  key: string
  label: string
  value: string
  desc: string
}

const userStore = useUserStore()
const settingStore = useSettingStore()

const loading = ref(true)
const echoTotal = ref<number | null>(null)
const unreadInboxCount = ref<number | null>(null)
const connectCount = ref<number | null>(null)

const username = computed(() => userStore.user?.username || '朋友')

const greeting = computed(() => {
  const hour = new Date().getHours()
  if (hour < 6) return '凌晨好'
  if (hour < 12) return '早上好'
  if (hour < 18) return '下午好'
  return '晚上好'
})

const dateText = computed(() => {
  return new Intl.DateTimeFormat('zh-CN', {
    month: 'long',
    day: 'numeric',
    weekday: 'long',
  }).format(new Date())
})

const dashboardStats = computed<StatCard[]>(() => {
  return [
    {
      key: 'echos',
      label: 'Echo 总数',
      value: echoTotal.value === null ? '--' : String(echoTotal.value),
      desc: '当前站点累计内容数量',
    },
    {
      key: 'inbox',
      label: '未读收件箱',
      value: unreadInboxCount.value === null ? '--' : String(unreadInboxCount.value),
      desc: '等待处理的系统消息',
    },
    {
      key: 'connect',
      label: '已连接节点',
      value: connectCount.value === null ? '--' : String(connectCount.value),
      desc: 'Connect 中已配置的远端',
    },
    {
      key: 'version',
      label: '当前版本',
      value: settingStore.hello?.version || '--',
      desc: '来自服务端 hello 信息',
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
    <section class="welcome-card">
      <p class="welcome-badge">{{ greeting }}, {{ username }}</p>
      <h2 class="welcome-title">欢迎来到 Ech0 Dashboard</h2>
      <p class="welcome-subtitle">今天是 {{ dateText }}，这里为你汇总当前站点的关键信息。</p>
    </section>

    <section class="stats-grid">
      <article v-for="item in dashboardStats" :key="item.key" class="stat-card">
        <p class="stat-label">{{ item.label }}</p>
        <p class="stat-value" :class="{ 'is-loading': loading && item.value === '--' }">{{ item.value }}</p>
        <p class="stat-desc">{{ item.desc }}</p>
      </article>
    </section>

  </div>
</template>

<style scoped>
.dashboard-page {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  width: 100%;
  padding: 0.25rem 0.5rem 0.75rem;
}

.welcome-card,
.stat-card {
  border: 1px solid var(--panel-border-soft);
  border-radius: var(--panel-radius-lg);
  background: var(--panel-surface-1);
  box-shadow: var(--panel-shadow-sm);
}

.welcome-card {
  padding: 1.2rem 1.25rem;
}

.welcome-badge {
  display: inline-block;
  padding: 0.2rem 0.65rem;
  border-radius: var(--panel-radius-sm);
  background: var(--panel-accent-weak);
  color: var(--panel-text-secondary);
  font-size: 0.95rem;
  font-weight: 700;
}

.welcome-title {
  margin-top: 0.65rem;
  font-size: 1.55rem;
  line-height: 1.3;
  color: var(--panel-text-primary);
  font-weight: 800;
  font-family: var(--font-display);
}

.welcome-subtitle {
  margin-top: 0.35rem;
  color: var(--panel-text-muted);
  font-size: 0.95rem;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(1, minmax(0, 1fr));
  gap: 0.85rem;
}

.stat-card {
  padding: 1rem 1.05rem;
  transition: transform 0.2s ease, border-color 0.2s ease, background-color 0.2s ease;
}

.stat-card:hover {
  transform: translateY(-1px);
  border-color: var(--panel-border-strong);
  background: var(--panel-surface-2);
}

.stat-label {
  font-size: 0.88rem;
  color: var(--panel-text-muted);
}

.stat-value {
  margin-top: 0.25rem;
  font-size: 1.5rem;
  line-height: 1.2;
  color: var(--panel-accent);
  font-weight: 800;
  font-family: var(--font-display);
}

.stat-value.is-loading {
  opacity: 0.7;
}

.stat-desc {
  margin-top: 0.28rem;
  font-size: 0.86rem;
  color: var(--panel-text-muted);
}

@media (min-width: 768px) {
  .dashboard-page {
    gap: 1.1rem;
  }

  .welcome-card {
    padding: 1.4rem 1.45rem;
  }

  .stats-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
