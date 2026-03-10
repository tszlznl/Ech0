<template>
  <div class="w-full px-2">
    <div class="mb-3 flex flex-col md:flex-row gap-2 md:items-center">
      <select
        v-model="level"
        class="h-9 rounded-[var(--panel-radius-md)] px-2 bg-[var(--input-bg-color)] text-[var(--panel-text-secondary)] border border-[var(--panel-border-soft)]"
      >
        <option value="all">全部级别</option>
        <option value="debug">debug</option>
        <option value="info">info</option>
        <option value="warn">warn</option>
        <option value="error">error</option>
      </select>
      <input
        v-model="keyword"
        placeholder="关键词过滤"
        class="h-9 rounded-[var(--panel-radius-md)] px-2 bg-[var(--input-bg-color)] text-[var(--panel-text-secondary)] border border-[var(--panel-border-soft)]"
      />
      <input
        v-model.number="tail"
        type="number"
        min="50"
        max="1000"
        class="h-9 w-24 rounded-[var(--panel-radius-md)] px-2 bg-[var(--input-bg-color)] text-[var(--panel-text-secondary)] border border-[var(--panel-border-soft)]"
      />
      <button class="h-9 px-3 rounded-md log-btn" @click="reload">应用过滤</button>
      <button class="h-9 px-3 rounded-md log-btn" @click="clearLogs">清屏</button>
      <label class="inline-flex items-center gap-1 text-sm text-[var(--panel-text-secondary)]">
        <input v-model="autoScroll" type="checkbox" />
        自动滚动
      </label>
      <span class="text-xs text-[var(--panel-text-muted)]">连接: {{ connectionText }}</span>
    </div>

    <div ref="logContainer" class="log-container text-xs md:text-sm">
      <template v-if="logs.length > 0">
        <div v-for="(line, idx) in logs" :key="`${line.time}-${idx}`" class="log-line">
          <span class="time">[{{ line.time || '-' }}]</span>
          <span class="level">[{{ line.level || 'info' }}]</span>
          <span class="msg">{{ line.msg }}</span>
        </div>
      </template>
      <div v-else class="text-[var(--panel-text-muted)]">暂无日志</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref } from 'vue'
import { fetchSystemLogs } from '@/service/api'
import { useOWebSocket } from '@/service/request/websocket'
import { getApiUrl, getWsUrl } from '@/service/request/shared'
import { theToast } from '@/utils/toast'

const logs = ref<App.Api.SystemLog.Entry[]>([])
const tail = ref(200)
const level = ref('all')
const keyword = ref('')
const autoScroll = ref(true)
const logContainer = ref<HTMLElement>()
const transport = ref<'ws' | 'sse'>('ws')
let es: EventSource | null = null

const { onMessage, open, close, status } = useOWebSocket<App.Api.Response<App.Api.SystemLog.Entry>>({
  url: getWsUrl('/ws/system/logs'),
  autoReconnect: {
    retries: 5,
    delay: 1000,
    onFailed: () => {
      startSSE()
    },
  },
  // 与后端协议对齐，避免 heartbeat ping/pong 不一致导致反复断连
  heartbeat: false,
})

const connectionText = computed(() => {
  if (transport.value === 'sse') return 'sse-fallback'
  return String(status.value)
})

const normalizeLevel = (value: string) => value.trim().toLowerCase()

const hitFilter = (entry: App.Api.SystemLog.Entry) => {
  const currentLevel = normalizeLevel(level.value)
  const currentKeyword = keyword.value.trim().toLowerCase()
  if (currentLevel !== 'all' && normalizeLevel(entry.level || '') !== currentLevel) {
    return false
  }
  if (!currentKeyword) {
    return true
  }
  const raw = `${entry.msg || ''} ${entry.raw || ''}`.toLowerCase()
  return raw.includes(currentKeyword)
}

const scrollToBottom = async () => {
  if (!autoScroll.value) return
  await nextTick()
  if (logContainer.value) {
    logContainer.value.scrollTop = logContainer.value.scrollHeight
  }
}

const pushLog = async (entry: App.Api.SystemLog.Entry) => {
  if (!hitFilter(entry)) return
  logs.value.push(entry)
  if (logs.value.length > 1000) {
    logs.value = logs.value.slice(logs.value.length - 1000)
  }
  await scrollToBottom()
}

const loadHistory = async () => {
  const res = await fetchSystemLogs({
    tail: tail.value,
    level: level.value,
    keyword: keyword.value,
  })
  if (res.code !== 1) {
    return
  }
  logs.value = Array.isArray(res.data) ? res.data : []
  await scrollToBottom()
}

const reload = async () => {
  await loadHistory()
  restartStream()
}

const clearLogs = () => {
  logs.value = []
}

const buildSSEUrl = () => {
  const token = localStorage.getItem('token')?.replace(/^"|"$/g, '') || ''
  const query = new URLSearchParams()
  query.set('token', token)
  if (level.value !== 'all') {
    query.set('level', level.value)
  }
  const trimmedKeyword = keyword.value.trim()
  if (trimmedKeyword) {
    query.set('keyword', trimmedKeyword)
  }
  const base = getApiUrl().replace(/\/+$/, '')
  return `${base}/system/logs/stream?${query.toString()}`
}

const stopSSE = () => {
  if (es) {
    es.close()
    es = null
  }
}

const startSSE = () => {
  if (transport.value === 'sse') return
  stopSSE()
  close()
  transport.value = 'sse'
  es = new EventSource(buildSSEUrl())
  es.onmessage = async (event) => {
    try {
      const payload = JSON.parse(event.data) as App.Api.Response<App.Api.SystemLog.Entry>
      if (payload.code === 1 && payload.data) {
        await pushLog(payload.data)
      }
    } catch {
      theToast.error('SSE 日志解析失败')
    }
  }
  es.onerror = () => {
    stopSSE()
  }
}

const startWS = () => {
  transport.value = 'ws'
  open()
}

const restartStream = () => {
  stopSSE()
  close()
  startWS()
}

onMounted(async () => {
  await loadHistory()
  startWS()
  onMessage(async (payload) => {
    if (transport.value !== 'ws') return
    if (payload.code !== 1 || !payload.data) return
    try {
      await pushLog(payload.data)
    } catch {
      theToast.error('系统日志处理失败')
    }
  })
})

onUnmounted(() => {
  stopSSE()
  close()
})
</script>

<style scoped>
.log-container {
  height: 68vh;
  overflow-y: auto;
  border: 1px solid var(--panel-border-soft);
  border-radius: var(--panel-radius-lg);
  padding: 12px;
  background: var(--panel-surface-1);
  box-shadow: var(--panel-shadow-sm);
}

.log-line {
  font-family: var(--font-mono);
  line-height: 1.55;
  margin-bottom: 2px;
  word-break: break-word;
  color: var(--panel-text-secondary);
}

.time {
  color: var(--panel-text-muted);
}

.level {
  margin: 0 6px;
  color: var(--panel-accent);
}

.msg {
  color: var(--panel-text-primary);
}

.log-btn {
  border: 1px solid var(--panel-border-soft);
  color: var(--panel-text-secondary);
  background: var(--panel-surface-1);
  box-shadow: var(--panel-shadow-sm);
}

.log-btn:hover {
  background: var(--panel-surface-2);
  border-color: var(--panel-border-strong);
}
</style>
