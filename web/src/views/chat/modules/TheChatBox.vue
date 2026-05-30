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
            <div
              v-if="msg.content.length === 0 && isStreaming(idx)"
              class="thinking"
              :aria-label="t('chatPanel.send')"
            >
              <span class="thinking__dot" />
              <span class="thinking__dot" />
              <span class="thinking__dot" />
            </div>
            <div v-else class="answer" :class="{ 'answer--streaming': isStreaming(idx) }">
              <TheMdPreview :content="msg.content" />
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
import { ref, computed, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { chatStream } from '@/service/api'
import { theToast } from '@/utils/toast'

const { t } = useI18n()
const router = useRouter()

const input = ref<string>('')
const loading = ref<boolean>(false)
const messages = ref<App.Api.Chat.ChatMessage[]>([])
const scrollArea = ref<HTMLElement | null>(null)
const inputEl = ref<HTMLTextAreaElement | null>(null)
let abort: (() => void) | null = null

const canSend = computed<boolean>(() => !loading.value && input.value.trim().length > 0)

// 是否正在流式输出最后一条 assistant 消息
const isStreaming = (idx: number): boolean =>
  loading.value && idx === messages.value.length - 1

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
  const assistant = ref<App.Api.Chat.ChatMessage>({ role: 'assistant', content: '', sources: [] })
  messages.value.push(assistant.value)
  input.value = ''
  loading.value = true
  nextTick(autoGrow)
  scrollToBottom()

  abort = chatStream(q, {
    onSources: (sources) => {
      assistant.value.sources = sources
      scrollToBottom()
    },
    onDelta: (text) => {
      assistant.value.content += text
      scrollToBottom()
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

// 流式进行中点击停止：中断请求但保留已生成的内容
const handleStop = () => {
  if (abort) abort()
  abort = null
  loading.value = false
}

const handleClear = () => {
  if (abort) abort()
  abort = null
  loading.value = false
  messages.value = []
}
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
  top: 0;
  left: 0;
  right: 0;
  /* 在横线上方再留出一段呼吸距离，避免内容黏住输入框 */
  bottom: calc(25dvh + 3rem);
  display: flex;
  justify-content: center;
  overflow-y: auto;
  /* 顶部与底部都做渐隐：内容靠近边界时柔和淡出 */
  -webkit-mask-image: linear-gradient(
    to bottom,
    transparent 0,
    #000 4.5rem,
    #000 calc(100% - 2.5rem),
    transparent 100%
  );
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
  margin-top: auto;
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
  border-radius: 1.1rem 1.1rem 0.25rem 1.1rem;
  background: var(--color-accent-soft);
  color: var(--color-text-primary);
  font-size: 0.9rem;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-word;
}

/* AI 回答：markdown 正文 */
.answer {
  width: 100%;
  font-size: 1rem;
}

.answer :deep(.echo-markdown) {
  line-height: 1.8;
}

/* flowtoken 式的流式光标：纯 CSS 接在最后一个块级元素末尾，
   markdown 重渲染也不受影响，无需逐 token 包裹 span */
.answer--streaming :deep(.echo-markdown > :last-child)::after {
  content: '';
  display: inline-block;
  width: 0.48rem;
  height: 1.05em;
  margin-left: 0.18rem;
  vertical-align: text-bottom;
  background: var(--color-accent);
  border-radius: 1px;
  animation: caret-blink 1.1s steps(1) infinite;
}

@keyframes caret-blink {
  0%,
  100% {
    opacity: 1;
  }
  50% {
    opacity: 0;
  }
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

@media (max-width: 640px) {
  .transcript__inner {
    padding-left: 1.25rem;
    padding-right: 1.25rem;
  }
}
</style>
