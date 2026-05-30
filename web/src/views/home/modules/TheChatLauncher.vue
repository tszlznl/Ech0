<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<!--
  对话快捷输入框：上方一句问候 + 下方圆角输入 pill。
  背景用主题画布色、输入用主题 surface 色，文字一律 text-primary，
  因此深/浅/暖三套主题都能保证对比度（不再依赖固定暗遮罩）。
  打字 → Enter 进入 /chat，Shift+Enter 换行，Esc / 点击空白关闭。
-->
<template>
  <Teleport to="body">
    <Transition name="chat-launcher">
      <div
        v-if="modelValue"
        class="chat-launcher"
        role="dialog"
        aria-modal="true"
        :aria-label="t('chatLauncher.title')"
        @mousedown.self="close"
      >
        <div class="chat-launcher__stage" @mousedown.self="close">
          <p class="chat-launcher__greeting">{{ t('chatLauncher.greeting') }}</p>

          <div class="chat-launcher__pill">
            <Chat class="chat-launcher__pill-icon" />
            <textarea
              ref="inputRef"
              v-model="draft"
              class="chat-launcher__input"
              rows="1"
              :placeholder="t('chatLauncher.placeholder')"
              :aria-label="t('chatLauncher.title')"
              @input="autoGrow"
              @keydown="onKeydown"
            />
            <!-- 点击发送：移动端无法靠回车提交时的主入口；桌面保留 Enter -->
            <button
              type="button"
              class="chat-launcher__send"
              :class="{ 'chat-launcher__send--active': draft.trim().length > 0 }"
              :aria-label="t('chatPanel.send')"
              @click="submit"
            >
              <span class="chat-launcher__send-glyph" aria-hidden="true">↑</span>
            </button>
          </div>

          <p class="chat-launcher__hint" aria-hidden="true">
            <kbd>↵</kbd>{{ t('chatLauncher.kbdEnter') }}
            <span class="chat-launcher__hint-sep">·</span>
            <kbd>esc</kbd>{{ t('chatLauncher.kbdClose') }}
          </p>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { nextTick, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import Chat from '@/components/icons/chat.vue'

const props = defineProps<{
  modelValue: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
}>()

const { t } = useI18n()
const router = useRouter()

const draft = ref<string>('')
const inputRef = ref<HTMLTextAreaElement | null>(null)

const close = () => {
  emit('update:modelValue', false)
}

const autoGrow = () => {
  const el = inputRef.value
  if (!el) return
  el.style.height = 'auto'
  el.style.height = `${Math.min(el.scrollHeight, window.innerHeight * 0.3)}px`
}

// 回车进入 /chat：带文本则透传 q 供聊天页自动发送，空文本则只打开对话页
const submit = () => {
  const q = draft.value.trim()
  close()
  router.push({ name: 'chat', query: q.length > 0 ? { q } : undefined })
}

const onKeydown = (e: KeyboardEvent) => {
  if (e.key === 'Enter' && !e.shiftKey && !e.isComposing) {
    e.preventDefault()
    submit()
  }
}

watch(
  () => props.modelValue,
  (open) => {
    if (open) {
      draft.value = ''
      void nextTick(() => {
        inputRef.value?.focus()
        autoGrow()
      })
    }
  },
)
</script>

<style scoped>
/* 背景用主题画布色（略透 + 模糊），而非固定暗遮罩：文字与底色同主题，恒定可读 */
.chat-launcher {
  position: fixed;
  inset: 0;
  z-index: 9000;
  background: color-mix(in srgb, var(--color-bg-canvas) 86%, transparent);
  backdrop-filter: blur(10px);
}

.chat-launcher__stage {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: flex-start;
  padding: clamp(6rem, 22vh, 12rem) 1.5rem 1.5rem;
  gap: 1.6rem;
}

.chat-launcher__greeting {
  margin: 0;
  font-size: clamp(1.4rem, 3vw, 1.9rem);
  font-weight: 600;
  letter-spacing: -0.01em;
  text-align: center;
  color: var(--color-text-primary);
}

/* 圆角输入 pill：自带 surface 底色 + 边框 + 柔和投影 */
.chat-launcher__pill {
  display: flex;

  /* 贴底对齐：多行时控件跟随最后一行落到底部；
     单行时控件高度 = 一行行高，视觉上即居中 */
  align-items: flex-end;
  gap: 0.65rem;
  width: min(40rem, 100%);
  padding: 0.7rem 1.15rem;
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border-subtle);
  border-radius: 1.6rem;
  box-shadow:
    0 1px 2px rgb(0 0 0 / 6%),
    0 14px 36px -16px rgb(0 0 0 / 26%);
  transition:
    border-color 0.18s ease,
    box-shadow 0.18s ease;
}

.chat-launcher__pill-icon {
  width: 1.2rem;
  height: 1.2rem;

  /* (一行 2rem − 图标 1.2rem) / 2，使单行时垂直居中、多行时贴底对齐基线 */
  margin-bottom: 0.4rem;
  color: var(--color-text-muted);
  flex: none;
}

:deep(.chat-launcher__pill-icon path) {
  fill: currentColor;
}

.chat-launcher__input {
  flex: 1 1 auto;
  min-width: 0;
  resize: none;
  max-height: 30vh;
  padding: 0.2rem 0;
  font-size: 1rem;
  line-height: 1.6;
  color: var(--color-text-primary);
  background: transparent;
  border: 0;
  outline: none;
  caret-color: var(--color-accent, #e07020);
  overflow-y: auto;
}

.chat-launcher__input::placeholder {
  color: var(--color-text-muted);
}

/* 发送按钮：圆形，有文字时点亮为 accent；移动端的主提交入口 */
.chat-launcher__send {
  flex: none;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2rem;
  height: 2rem;
  border: 0;
  border-radius: 999px;
  background: var(--color-bg-muted);
  color: var(--color-text-muted);
  cursor: pointer;
  transition:
    color 0.15s ease,
    background 0.15s ease,
    transform 0.12s ease,
    filter 0.15s ease;
}

.chat-launcher__send--active {
  background: var(--color-accent, #e07020);
  color: #fff;
}

.chat-launcher__send:hover {
  transform: translateY(-1px);
  filter: brightness(1.03);
}

.chat-launcher__send:active {
  transform: translateY(0);
}

.chat-launcher__send-glyph {
  font-size: 1.05rem;
  line-height: 1;
  font-weight: 700;
  font-variant-emoji: text;
}

/* 极淡的一行提示，低存在感 */
.chat-launcher__hint {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  margin: 0;
  font-size: 0.72rem;
  color: var(--color-text-muted);
  opacity: 0.7;
  letter-spacing: 0.01em;
}

.chat-launcher__hint kbd {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 1.15rem;
  height: 1.15rem;
  margin-right: 0.25rem;
  padding: 0 0.28rem;
  font-family: var(--font-family-mono, monospace);
  font-size: 0.66rem;
  line-height: 1;
  color: var(--color-text-secondary);
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border-subtle);
  border-radius: 4px;
}

.chat-launcher__hint-sep {
  opacity: 0.6;
}

@media (width <= 640px) {
  .chat-launcher__stage {
    padding-top: 4.5rem;
  }

  .chat-launcher__hint {
    display: none;
  }
}

.chat-launcher-enter-active,
.chat-launcher-leave-active {
  transition: opacity 0.2s ease;
}

.chat-launcher-enter-active .chat-launcher__stage,
.chat-launcher-leave-active .chat-launcher__stage {
  transition:
    transform 0.24s cubic-bezier(0.2, 0.9, 0.3, 1.1),
    opacity 0.24s ease;
}

.chat-launcher-enter-from,
.chat-launcher-leave-to {
  opacity: 0;
}

.chat-launcher-enter-from .chat-launcher__stage,
.chat-launcher-leave-to .chat-launcher__stage {
  opacity: 0;
  transform: translateY(-10px);
}
</style>
