<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="hairline-chat">
    <!-- 角落控件：无边框幽灵图标 -->
    <button class="ghost-ctrl ghost-ctrl--back" :title="t('commonNav.backHome')" @click="goHome">
      <Back class="ghost-ctrl__icon" />
    </button>
    <button
      v-if="messages.length > 0"
      class="ghost-ctrl ghost-ctrl--clear"
      :title="t('chatPanel.clear')"
      @click="handleClear"
    >
      <Close class="ghost-ctrl__icon" />
    </button>

    <!-- 对话区：从线的上方向上生长，贴底排列 -->
    <div ref="scrollArea" class="transcript">
      <div class="transcript__inner">
        <div
          v-for="(msg, idx) in messages"
          :key="idx"
          class="turn"
          :class="msg.role === 'user' ? 'turn--user' : 'turn--ai'"
        >
          <!-- 用户：柔和的 accent-soft 气泡，右对齐 -->
          <p v-if="msg.role === 'user'" class="bubble">{{ msg.content }}</p>

          <!-- AI：无气泡的 markdown 正文 -->
          <template v-else>
            <!-- 检索状态条：Agent 自主检索时逐条显示关键词（设计 §9 searching 事件） -->
            <div v-if="msg.searches && msg.searches.length > 0" class="searching">
              <span
                v-for="(query, qi) in msg.searches"
                :key="qi"
                class="searching__chip"
                :class="{
                  'searching__chip--live': isStreaming(idx) && qi === msg.searches.length - 1,
                }"
              >
                {{ t('chatPanel.searching', { query }) }}
              </span>
            </div>

            <div
              v-if="msg.content.length === 0 && isStreaming(idx)"
              class="thinking"
              :aria-label="t('chatPanel.send')"
            >
              <span class="thinking__dot" />
              <span class="thinking__dot" />
              <span class="thinking__dot" />
            </div>
            <div v-else class="answer">
              <!-- 流式 + 揭示未追平时都用逐 token 动画；等揭示真正播完再切到带复制/折叠的完整渲染器，
                   避免慢节奏下尾巴整坨弹出 -->
              <AnimatedMarkdown
                v-if="showAnimated(idx)"
                :content="msg.content"
                :streaming="isStreaming(idx)"
                animation="blurIn"
                @update:revealing="assistantRevealing = $event"
              />
              <TheMdPreview v-else :content="msg.content" />
            </div>
          </template>

          <!-- 引用来源：极小的 accent 文字链，无边框 -->
          <div v-if="msg.sources && msg.sources.length > 0" class="sources">
            <button
              v-for="src in msg.sources"
              :key="src.echo_id"
              class="sources__link"
              @click="goToEcho(src.echo_id)"
            >
              <span class="sources__mark">↗</span>{{ formatSource(src) }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 输入区：textarea 的下边框就是那条横线（钉在 75vh） -->
    <div class="composer" :class="{ 'composer--active': input.trim().length > 0 || loading }">
      <textarea
        ref="inputEl"
        v-model="input"
        class="composer__field"
        rows="1"
        :placeholder="t('chatPanel.inputPlaceholder')"
        @input="autoGrow"
        @keydown="handleKeydown"
      />
      <Transition name="send-pop">
        <button
          v-if="loading"
          class="composer__action composer__action--stop"
          :title="t('chatPanel.clear')"
          @click="handleStop"
        >
          <span class="composer__stop-glyph" />
        </button>
        <button
          v-else-if="canSend"
          class="composer__action composer__action--send"
          :title="t('chatPanel.send')"
          @click="send(input)"
        >
          <span class="composer__return-glyph">↩︎</span>
        </button>
      </Transition>
    </div>

    <!-- 线下留白区：空态放安静的预设问题，否则放一句说明 -->
    <div class="understory">
      <template v-if="messages.length === 0">
        <p class="understory__hint">{{ t('chatPanel.suggestionsTitle') }}</p>
        <button
          v-for="(s, i) in suggestions"
          :key="i"
          class="understory__suggestion"
          @click="send(s)"
        >
          {{ s }}
        </button>
      </template>
      <p v-else class="understory__intro">{{ t('chatPanel.intro') }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import Back from '@/components/icons/back.vue'
import Close from '@/components/icons/close.vue'
import { TheMdPreview } from '@/components/advanced/md'
import AnimatedMarkdown from './AnimatedMarkdown.vue'
import { ref, computed, nextTick, onBeforeUnmount, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { chatStream } from '@/service/api'
import { getChatSession, clearChatSession } from '@/service/api/chat'
import { useBaseDialog } from '@/composables/useBaseDialog'
import { theToast } from '@/utils/toast'

const { t } = useI18n()
const router = useRouter()
const { openConfirm } = useBaseDialog()

const input = ref<string>('')
const loading = ref<boolean>(false)
const messages = ref<App.Api.Chat.ChatMessage[]>([])
const scrollArea = ref<HTMLElement | null>(null)
const inputEl = ref<HTMLTextAreaElement | null>(null)
let abort: (() => void) | null = null

const canSend = computed<boolean>(() => !loading.value && input.value.trim().length > 0)

// 最后一条 assistant 消息的逐 token 揭示是否尚未追平（由 AnimatedMarkdown 上报）
const assistantRevealing = ref<boolean>(false)

// 是否正在流式输出最后一条 assistant 消息
const isStreaming = (idx: number): boolean => loading.value && idx === messages.value.length - 1

// 是否对该消息使用动画渲染：流式中、或流已结束但揭示还没播完，都继续动画
const showAnimated = (idx: number): boolean => {
  const last = messages.value.length - 1
  if (idx !== last || messages.value[idx]?.role !== 'assistant') return false
  return loading.value || assistantRevealing.value
}

const suggestions = computed<string[]>(() => [
  t('chatPanel.suggestion1'),
  t('chatPanel.suggestion2'),
  t('chatPanel.suggestion3'),
])

const scrollToBottom = () => {
  nextTick(() => {
    if (scrollArea.value) {
      scrollArea.value.scrollTop = scrollArea.value.scrollHeight
    }
  })
}

// 流式期间逐帧粘住底部，跟随节流揭示平滑滚动；用户上滑离开底部则停止打扰
let stickRaf = 0
const isNearBottom = (): boolean => {
  const el = scrollArea.value
  if (!el) return true
  return el.scrollHeight - el.scrollTop - el.clientHeight < 80
}
const stickToBottom = () => {
  const el = scrollArea.value
  if (el && isNearBottom()) el.scrollTop = el.scrollHeight
  // 流式中、以及流结束后揭示仍在播放时，都保持粘底
  const keepGoing = loading.value || assistantRevealing.value
  stickRaf =
    keepGoing && typeof requestAnimationFrame === 'function'
      ? requestAnimationFrame(stickToBottom)
      : 0
}
const stopStick = () => {
  if (stickRaf) cancelAnimationFrame(stickRaf)
  stickRaf = 0
}

// textarea 向上生长，横线恒定钉在 75vh
const autoGrow = () => {
  const el = inputEl.value
  if (!el) return
  el.style.height = 'auto'
  el.style.height = `${Math.min(el.scrollHeight, window.innerHeight * 0.28)}px`
}

const formatSource = (src: App.Api.Chat.ChatSource): string => {
  const day = new Date(src.echo_created * 1000).toISOString().slice(0, 10)
  const text = src.content.length > 40 ? src.content.slice(0, 40) + '…' : src.content
  return ` ${day} · ${text}`
}

const goHome = () => router.push('/')
const goToEcho = (echoId: string) => router.push(`/echo/${echoId}`)

const send = (question: string) => {
  const q = question.trim()
  if (q.length === 0 || loading.value) return

  messages.value.push({ role: 'user', content: q })
  const assistant = ref<App.Api.Chat.ChatMessage>({
    role: 'assistant',
    content: '',
    sources: [],
    searches: [],
  })
  messages.value.push(assistant.value)
  input.value = ''
  loading.value = true
  assistantRevealing.value = true
  nextTick(autoGrow)
  scrollToBottom()
  stickToBottom()

  abort = chatStream(q, {
    onSearching: (query) => {
      if (query && !assistant.value.searches?.includes(query)) {
        assistant.value.searches?.push(query)
      }
    },
    onSources: (sources) => {
      // sources 可多次增量到达，按 echo_id 累积去重（设计 §9）
      const merged = assistant.value.sources ? [...assistant.value.sources] : []
      const seen = new Set(merged.map((s) => s.echo_id))
      for (const src of sources) {
        if (!seen.has(src.echo_id)) {
          seen.add(src.echo_id)
          merged.push(src)
        }
      }
      assistant.value.sources = merged
    },
    onDelta: (text) => {
      assistant.value.content += text
    },
    onError: (message) => {
      loading.value = false
      theToast.error(message || String(t('chatPanel.errorGeneric')))
      if (assistant.value.content.length === 0) {
        assistant.value.content = String(t('chatPanel.errorGeneric'))
      }
    },
    onDone: () => {
      loading.value = false
    },
  })
}

const handleKeydown = (e: KeyboardEvent) => {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    send(input.value)
  }
}

// 流式进行中点击停止：中断请求但保留已生成的内容（剩余尾巴让揭示从容播完）
const handleStop = () => {
  if (abort) abort()
  abort = null
  loading.value = false
}

const handleClear = () => {
  openConfirm({
    title: t('chatPanel.clearConfirmTitle'),
    description: t('chatPanel.clearConfirmDesc'),
    onConfirm: async () => {
      if (abort) abort()
      abort = null
      loading.value = false
      assistantRevealing.value = false
      stopStick()
      try {
        await clearChatSession()
      } catch {
        // 清除失败不阻断本地清空（best-effort）
      }
      messages.value = []
      theToast.success(String(t('chatPanel.clearSuccess')))
    },
  })
}

// 进入页面恢复上次的持久化会话（仅展示）。恢复的消息走静态渲染：
// loading 与 assistantRevealing 均为 false，showAnimated 对历史消息返回 false → 走 TheMdPreview。
onMounted(async () => {
  try {
    const res = await getChatSession()
    const history = res.data
    if (Array.isArray(history) && history.length > 0) {
      messages.value = history
      scrollToBottom()
    }
  } catch {
    // 恢复失败静默忽略，保持空态
  }
})

onBeforeUnmount(stopStick)
</script>

<style scoped>
.hairline-chat {
  position: relative;
  width: 100%;
  height: 100vh;
  height: 100dvh;
  overflow: hidden;
  background: var(--color-bg-canvas);
  color: var(--color-text-primary);
}

/* ── 角落幽灵控件 ───────────────────────────── */
.ghost-ctrl {
  position: absolute;
  top: 1.25rem;
  z-index: 3;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2.25rem;
  height: 2.25rem;
  border: none;
  background: transparent;
  color: var(--color-text-muted);
  cursor: pointer;
  opacity: 0.45;
  transition:
    opacity 0.2s ease,
    color 0.2s ease;
}

.ghost-ctrl--back {
  left: 1.25rem;
}

.ghost-ctrl--clear {
  right: 1.25rem;
}

.ghost-ctrl:hover {
  opacity: 1;
  color: var(--color-text-primary);
}

.ghost-ctrl__icon {
  width: 1.2rem;
  height: 1.2rem;
}

/* 图标内置 fill=#888888，统一改用 currentColor 以便随主题/状态着色 */
.ghost-ctrl__icon :deep(path) {
  fill: currentColor;
}

/* ── 对话区（线之上，贴底向上生长） ─────────── */
.transcript {
  position: absolute;
  inset: 0 0 calc(25dvh + 3rem);

  /* 在横线上方再留出一段呼吸距离，避免内容黏住输入框 */
  display: flex;
  justify-content: center;
  overflow-y: auto;

  /* 顶部与底部都做渐隐：内容靠近边界时柔和淡出 */
  mask-image: linear-gradient(
    to bottom,
    transparent 0,
    #000 4.5rem,
    #000 calc(100% - 2.5rem),
    transparent 100%
  );
}

.transcript__inner {
  width: 100%;
  max-width: 42rem;

  /* 顶对齐、向下生长：内容不足时不再贴底，避免每多一行整块往上跳一格；
     溢出后由 scrollToBottom 粘住底部 */
  align-self: flex-start;
  padding: 4.5rem 1.5rem 0.5rem;
  display: flex;
  flex-direction: column;
  gap: 1.6rem;
}

.turn {
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
  animation: turn-in 0.32s ease both;
}

.turn--user {
  align-items: flex-end;
}

.turn--ai {
  align-items: flex-start;
}

@keyframes turn-in {
  from {
    opacity: 0;
    transform: translateY(6px);
  }

  to {
    opacity: 1;
    transform: none;
  }
}

/* 用户气泡：柔和 accent-soft，无硬边框 */
.bubble {
  max-width: 85%;
  padding: 0.55rem 0.9rem;
  border-radius: 1.1rem 1.1rem 0.25rem;
  background: var(--color-accent-soft);
  color: var(--color-text-primary);
  font-size: 0.9rem;
  line-height: 1.6;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}

/* AI 回答：markdown 正文 */
.answer {
  width: 100%;
  font-size: 1rem;
}

.answer :deep(.echo-markdown) {
  line-height: 1.8;
}

/* 首 token 到达前的「思考中」动画 */
.thinking {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.4rem 0;
}

.thinking__dot {
  width: 0.4rem;
  height: 0.4rem;
  border-radius: 999px;
  background: var(--color-text-muted);
  opacity: 0.5;
  animation: thinking-bounce 1.2s ease-in-out infinite;
}

.thinking__dot:nth-child(2) {
  animation-delay: 0.16s;
}

.thinking__dot:nth-child(3) {
  animation-delay: 0.32s;
}

@keyframes thinking-bounce {
  0%,
  80%,
  100% {
    transform: translateY(0);
    opacity: 0.35;
  }

  40% {
    transform: translateY(-0.28rem);
    opacity: 1;
  }
}

/* ── 检索状态条（Agent 自主检索） ───────────── */
.searching {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
  margin-bottom: 0.5rem;
}

.searching__chip {
  display: inline-flex;
  align-items: center;
  max-width: 22rem;
  padding: 0.1rem 0.5rem;
  border-radius: 999px;
  background: var(--color-accent-soft);
  color: var(--color-text-secondary);
  font-size: 0.72rem;
  line-height: 1.5;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.searching__chip::before {
  content: '🔍';
  margin-right: 0.3rem;
  font-size: 0.7rem;
  font-variant-emoji: text;
}

/* 最新一条仍在检索：轻微呼吸 */
.searching__chip--live {
  animation: searching-pulse 1.4s ease-in-out infinite;
}

@keyframes searching-pulse {
  0%,
  100% {
    opacity: 0.55;
  }

  50% {
    opacity: 1;
  }
}

/* ── 引用来源 ───────────────────────────────── */
.sources {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
  margin-top: 0.35rem;
}

.sources__link {
  display: inline-flex;
  align-items: baseline;
  max-width: 32rem;
  border: none;
  background: transparent;
  padding: 0;
  font-size: 0.72rem;
  line-height: 1.5;
  color: var(--color-text-muted);
  text-align: left;
  cursor: pointer;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  transition: color 0.18s ease;
}

.sources__link:hover {
  color: var(--color-accent);
}

.sources__mark {
  color: var(--color-accent);
  opacity: 0.8;
}

/* ── 输入区：下边框就是 75vh 那条横线 ───────── */
.composer {
  position: absolute;
  left: 50%;
  bottom: 25vh;
  bottom: 25dvh;
  transform: translateX(-50%);
  z-index: 2;
  width: min(42rem, calc(100% - 3rem));
  display: flex;
  align-items: flex-end;
  gap: 0.6rem;
  border-bottom: 1px solid var(--color-border-strong);
  transition: border-color 0.25s ease;
}

.composer--active,
.composer:focus-within {
  border-bottom-color: var(--color-accent);
}

.composer__field {
  flex: 1;
  resize: none;
  border: none;
  outline: none;
  background: transparent;
  padding: 0 0 0.7rem;
  max-height: 28vh;
  font-size: 1rem;
  line-height: 1.6;
  color: var(--color-text-primary);
  overflow-y: auto;
}

.composer__field::placeholder {
  color: var(--color-text-muted);
  opacity: 0.7;
}

/* 发送 / 停止：仅在可操作时出现，一个小小的极简字形，不抢戏 */
.composer__action {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 1.6rem;
  height: 1.6rem;
  margin-bottom: 0.5rem;
  border: none;
  background: transparent;
  color: var(--color-text-muted);
  cursor: pointer;
  transition:
    color 0.18s ease,
    transform 0.18s ease;
}

.composer__action--send {
  color: var(--color-accent);
}

.composer__action:hover {
  transform: translateY(-1px);
  filter: brightness(1.1);
}

.composer__action:active {
  transform: translateY(0);
}

.composer__return-glyph {
  font-size: 1.05rem;
  line-height: 1;

  /* 强制文本字形，避免被渲染成 emoji 样式 */
  font-variant-emoji: text;
}

.composer__stop-glyph {
  width: 0.6rem;
  height: 0.6rem;
  border-radius: 0.14rem;
  background: currentColor;
}

/* 发送按钮弹入/弹出 */
.send-pop-enter-active,
.send-pop-leave-active {
  transition:
    transform 0.18s cubic-bezier(0.34, 1.56, 0.64, 1),
    opacity 0.18s ease;
}

.send-pop-enter-from,
.send-pop-leave-to {
  transform: scale(0.5);
  opacity: 0;
}

/* ── 线下留白区（25vh） ─────────────────────── */
.understory {
  position: absolute;
  left: 50%;
  top: 75vh;
  top: 75dvh;
  transform: translateX(-50%);
  width: min(42rem, calc(100% - 3rem));
  padding-top: 1.5rem;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.35rem;
}

.understory__hint {
  font-size: 0.72rem;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--color-text-muted);
  opacity: 0.7;
  margin-bottom: 0.25rem;
}

.understory__suggestion {
  border: none;
  background: transparent;
  padding: 0.15rem 0;
  font-size: 0.9rem;
  line-height: 1.5;
  color: var(--color-text-secondary);
  text-align: left;
  cursor: pointer;
  transition:
    color 0.18s ease,
    transform 0.18s ease;
}

.understory__suggestion:hover {
  color: var(--color-accent);
  transform: translateX(3px);
}

.understory__intro {
  font-size: 0.78rem;
  line-height: 1.65;
  color: var(--color-text-muted);
  opacity: 0.75;
  max-width: 34rem;
}

@media (width <= 640px) {
  .transcript__inner {
    padding-left: 1.25rem;
    padding-right: 1.25rem;
  }
}
</style>
