<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { fetchGetEchosByPage, fetchGetConnectList, fetchUnreadInbox } from '@/service/api'
import { useUserStore, useSettingStore } from '@/stores'

type StatCard = {
  key: string
  label: string
  value: string
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
    <section class="welcome-header">
      <div class="welcome-main">
        <h2 class="welcome-username">{{ username }}，欢迎回来 <span class="wave-hand">👋</span></h2>
        <p class="welcome-greeting">{{ greeting }}</p>
        <p class="welcome-tip">今天也来记录一点新的灵感</p>
      </div>
      <p class="welcome-date">{{ dateText }}</p>
    </section>

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
  display: flex;
  flex-direction: column;
  gap: 0.95rem;
  width: 100%;
  padding: 0.4rem 0.25rem 0.9rem;
}

.welcome-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 0.5rem 0.25rem 0.65rem;
  border-bottom: 1px solid var(--color-border-subtle);
}

.welcome-main {
  min-width: 0;
  text-align: left;
}

.welcome-greeting {
  font-size: 0.9rem;
  line-height: 1.2;
  color: var(--color-text-muted);
  font-weight: 600;
  margin-top: 0.35rem;
}

.welcome-username {
  margin: 0;
  font-size: clamp(1.45rem, 2.4vw, 1.9rem);
  line-height: 1.2;
  color: var(--color-text-primary);
  font-weight: 800;
  font-family: var(--font-family-display);
  letter-spacing: 0.01em;
}

.welcome-tip {
  margin-top: 0.3rem;
  color: var(--color-text-muted);
  font-size: 0.88rem;
}

.welcome-date {
  display: inline-flex;
  align-items: center;
  gap: 0.2rem;
  margin-top: 0.15rem;
  color: var(--color-text-secondary);
  font-size: 0.9rem;
  white-space: nowrap;
  font-weight: 600;
}

.wave-hand {
  display: inline-block;
  margin-left: 0.2rem;
  transform-origin: center;
  will-change: transform;
}

.wave-hand:hover {
  animation: hand-shake 620ms ease-in-out;
}

@keyframes hand-shake {
  0% {
    transform: rotate(0deg) scale(1);
  }
  15% {
    transform: rotate(16deg) scale(1.08);
  }
  30% {
    transform: rotate(-14deg) scale(1.08);
  }
  45% {
    transform: rotate(12deg) scale(1.06);
  }
  60% {
    transform: rotate(-10deg) scale(1.04);
  }
  75% {
    transform: rotate(7deg) scale(1.02);
  }
  100% {
    transform: rotate(0deg) scale(1);
  }
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
  .dashboard-page {
    gap: 1rem;
  }

  .stats-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 640px) {
  .welcome-header {
    align-items: flex-start;
    flex-direction: column;
    gap: 0.3rem;
  }

  .welcome-date {
    white-space: normal;
  }
}
</style>
