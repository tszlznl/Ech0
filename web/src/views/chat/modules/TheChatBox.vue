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
      <div ref="transcriptInner" class="transcript__inner">
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

            <!-- 覆盖度状态条：summarize_echos 区间聚合（年终/月度总结）时如实展示覆盖范围，杜绝静默截断 -->
            <div v-if="msg.coverage" class="searching">
              <span class="searching__chip">
                {{
                  msg.coverage.truncated
                    ? t('chatPanel.coverageTruncated', { returned: msg.coverage.returned })
                    : t('chatPanel.coverage', { total: msg.coverage.total })
                }}
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

          <!-- 引用来源：默认展示前三条，其余折叠（ChatSources 内部管理展开态） -->
          <ChatSources
            v-if="msg.sources && msg.sources.length > 0"
            :sources="msg.sources"
            @open="goToEcho"
          />
        </div>
      </div>
    </div>

    <!-- 输入区：textarea 的下边框就是那条横线（钉在 75vh） -->
    <div
      class="composer"
      :class="{
        'composer--active': input.trim().length > 0 || loading,
        'composer--loading': loading,
      }"
    >
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
          <Send class="composer__send-icon" />
        </button>
      </Transition>
    </div>

    <!-- 线下留白区：输入框为空时常驻预设场景，开始输入则淡出 -->
    <div class="understory">
      <Transition name="understory-fade">
        <div v-if="showSuggestions" class="understory__list">
          <p class="understory__hint">{{ t('chatPanel.suggestionsTitle') }}</p>
          <button
            v-for="(s, i) in suggestions"
            :key="i"
            class="understory__suggestion"
            @click="send(s)"
          >
            {{ s }}
          </button>
        </div>
      </Transition>
    </div>
  </div>
</template>

<script setup lang="ts">
import Back from '@/components/icons/back.vue'
import Close from '@/components/icons/close.vue'
import Send from '@/components/icons/send.vue'
import { TheMdPreview } from '@/components/advanced/md'
import AnimatedMarkdown from './AnimatedMarkdown.vue'
import ChatSources from './ChatSources.vue'
import { ref, computed, nextTick, onBeforeUnmount, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { chatStream } from '@/service/api'
import { getChatSession, clearChatSession } from '@/service/api/chat'
import { useBaseDialog } from '@/composables/useBaseDialog'
import { theToast } from '@/utils/toast'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const { openConfirm } = useBaseDialog()

const input = ref<string>('')
const loading = ref<boolean>(false)
const messages = ref<App.Api.Chat.ChatMessage[]>([])
const scrollArea = ref<HTMLElement | null>(null)
const transcriptInner = ref<HTMLElement | null>(null)
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
  t('chatPanel.suggestion4'),
])

// 输入框为空且非流式时展示预设场景；开始输入或发送后淡出
const showSuggestions = computed<boolean>(() => !loading.value && input.value.trim().length === 0)

// 贴底滚动：事件驱动而非帧驱动。`pinned` 是用户「想不想贴底」的意图，只由真实滚动翻转；
// 内容长高由 ResizeObserver 感知后跟随一次。彻底告别 rAF 每帧强写 scrollTop 带来的亚像素抖动。
const STICK_THRESHOLD = 80
const pinned = ref<boolean>(true)
let resizeObserver: ResizeObserver | null = null

const jumpToBottom = () => {
  nextTick(() => {
    const el = scrollArea.value
    if (el) el.scrollTop = el.scrollHeight
  })
}

// 用户滚动时更新贴底意图：离底超过阈值即视为「想自己翻看」，滚回阈值内则恢复跟随。
// 程序触发的「跳到底」滚完仍在底部 → 读出 pinned=true，不会误翻转，故无需额外守卫。
const onScroll = () => {
  const el = scrollArea.value
  if (!el) return
  pinned.value = el.scrollHeight - el.scrollTop - el.clientHeight < STICK_THRESHOLD
}

// 内容尺寸变化（逐 token 揭示、流结束揭示尾词、渲染器切换）→ 若意图贴底则跟随一次。
// 已在底部时把 scrollTop 设成它本来的值是 no-op，不触发 scroll 事件、不抖。
const onContentResize = () => {
  const el = scrollArea.value
  if (el && pinned.value) el.scrollTop = el.scrollHeight
}

// textarea 向上生长，横线恒定钉在 75vh
const autoGrow = () => {
  const el = inputEl.value
  if (!el) return
  el.style.height = 'auto'
  el.style.height = `${Math.min(el.scrollHeight, window.innerHeight * 0.28)}px`
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
  pinned.value = true
  nextTick(autoGrow)
  jumpToBottom()

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
    onCoverage: (coverage) => {
      // 区间聚合总结（summarize_echos）的覆盖度，供「📚 已覆盖 N 条」状态条如实展示
      assistant.value.coverage = coverage
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
      pinned.value = true
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
  // 贴底引擎接线：scroll 监听更新意图，ResizeObserver 感知内容长高后跟随
  const area = scrollArea.value
  if (area) area.addEventListener('scroll', onScroll, { passive: true })
  if (transcriptInner.value && typeof ResizeObserver === 'function') {
    resizeObserver = new ResizeObserver(onContentResize)
    resizeObserver.observe(transcriptInner.value)
  }

  try {
    const res = await getChatSession()
    const history = res.data
    if (Array.isArray(history) && history.length > 0) {
      messages.value = history
      pinned.value = true
      jumpToBottom()
    }
  } catch {
    // 恢复失败静默忽略，保持空态
  }

  // 由快捷输入框（Cmd/Ctrl+J）带入的问题：恢复历史后自动发送，并清掉 query 防止刷新重发
  const initialQuery = route.query.q
  const q = Array.isArray(initialQuery) ? initialQuery[0] : initialQuery
  if (typeof q === 'string' && q.trim().length > 0) {
    void router.replace({ query: {} })
    send(q)
  }
})

onBeforeUnmount(() => {
  resizeObserver?.disconnect()
  resizeObserver = null
  scrollArea.value?.removeEventListener('scroll', onScroll)
})
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
  border-radius: 999px;

  /* 磨砂圆底：无论背后滚动到什么文字都保持可辨，移动端窄屏尤甚 */
  background: color-mix(in srgb, var(--color-bg-canvas) 70%, transparent);
  backdrop-filter: blur(8px);
  color: var(--color-text-muted);
  cursor: pointer;
  opacity: 0.7;
  transition:
    opacity 0.2s ease,
    background 0.2s ease,
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
  background: color-mix(in srgb, var(--color-bg-canvas) 88%, transparent);
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

  /* 贴底滚动由 JS 手动管理：关掉浏览器原生锚定，避免两者抢着调 scrollTop 造成抖动 */
  overflow-anchor: none;

  /* 顶部与底部都做渐隐：内容靠近边界时柔和淡出 */
  mask-image: linear-gradient(
    to bottom,
    transparent 0,
    #000 4.5rem,
    #000 calc(100% - 1rem),
    transparent 100%
  );
}

.transcript__inner {
  width: 100%;
  max-width: 42rem;

  /* 顶对齐、向下生长：内容不足时不再贴底，避免每多一行整块往上跳一格；
     溢出后由 jumpToBottom + ResizeObserver 跟随粘住底部 */
  align-self: flex-start;

  /* 底部留出 > 1rem 的 gutter：贴底时让最后一条来源避开 .transcript 底部 1rem 的渐隐带，
     杜绝来源块被 mask 半透明笼罩、随滚动逐帧跳变透明度的闪烁 */
  padding: 4.5rem 1.5rem 1.5rem;
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
}

/* 发丝线（底色）：两端淡出的渐变，避免满宽硬边显得空荡 */
.composer::after {
  content: '';
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  height: 1px;
  background: linear-gradient(
    to right,
    transparent,
    var(--color-border-strong) 18%,
    var(--color-border-strong) 82%,
    transparent
  );
}

/* accent 线：聚焦/有内容时从中间向两端"画"出来，覆盖在灰线之上。
   两端淡出交给 mask，颜色/流光交给 background，互不干扰 */
.composer::before {
  content: '';
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 1;
  height: 1px;
  background: var(--color-accent);
  mask-image: linear-gradient(to right, transparent, #000 15%, #000 85%, transparent);
  transform: scaleX(0);
  transform-origin: 50% 50%;
  transition: transform 0.35s cubic-bezier(0.22, 1, 0.36, 1);
}

.composer--active::before,
.composer:focus-within::before {
  transform: scaleX(1);
}

/* AI 回复时让 accent 线轻微流光，把这条贯穿全页的线当作状态指示 */
.composer--loading::before {
  background-image: linear-gradient(
    90deg,
    var(--color-accent) 0%,
    color-mix(in srgb, var(--color-accent) 35%, #fff) 50%,
    var(--color-accent) 100%
  );
  background-size: 200% 100%;
  animation: composer-shimmer 1.6s linear infinite;
}

@keyframes composer-shimmer {
  from {
    background-position: 150% 0;
  }

  to {
    background-position: -150% 0;
  }
}

@media (prefers-reduced-motion: reduce) {
  .composer::before {
    transition: none;
  }

  .composer--loading::before {
    animation: none;
  }
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

.composer__send-icon {
  width: 1.15rem;
  height: 1.15rem;
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
}

.understory__list {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.35rem;
}

.understory-fade-enter-active,
.understory-fade-leave-active {
  transition:
    opacity 0.25s ease,
    transform 0.25s ease;
}

.understory-fade-enter-from,
.understory-fade-leave-to {
  opacity: 0;
  transform: translateY(4px);
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

@media (width <= 640px) {
  .transcript__inner {
    padding-left: 1.25rem;
    padding-right: 1.25rem;
  }
}
</style>
