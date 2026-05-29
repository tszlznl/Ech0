<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <PanelCard>
    <div class="w-full flex flex-col" style="min-height: 60vh">
      <div class="flex flex-row items-center justify-between mb-3">
        <h1 class="text-[var(--color-text-primary)] font-bold text-lg">
          {{ t('chatPanel.title') }}
        </h1>
        <BaseButton v-if="messages.length > 0" :icon="Delete" @click="handleClear">
          {{ t('chatPanel.clear') }}
        </BaseButton>
      </div>

      <p class="text-xs text-[var(--color-text-secondary)] opacity-70 mb-4">
        {{ t('chatPanel.intro') }}
      </p>

      <!-- 消息区 -->
      <div ref="scrollArea" class="flex-1 overflow-y-auto flex flex-col gap-4 pr-1">
        <!-- 空态：预设问题 -->
        <div v-if="messages.length === 0" class="flex flex-col gap-2">
          <p class="text-sm text-[var(--color-text-secondary)] font-semibold">
            {{ t('chatPanel.suggestionsTitle') }}
          </p>
          <button
            v-for="(s, i) in suggestions"
            :key="i"
            class="text-left px-3 py-2 rounded-[var(--radius-md)] border text-sm text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)] hover:bg-[var(--color-bg-muted)] transition-colors"
            :style="{ borderColor: 'var(--color-border-subtle)' }"
            @click="send(s)"
          >
            {{ s }}
          </button>
        </div>

        <!-- 对话 -->
        <div
          v-for="(msg, idx) in messages"
          :key="idx"
          class="flex flex-col gap-1"
          :class="msg.role === 'user' ? 'items-end' : 'items-start'"
        >
          <div
            class="max-w-[85%] px-3 py-2 rounded-[var(--radius-md)] text-sm whitespace-pre-wrap break-words"
            :class="
              msg.role === 'user'
                ? 'bg-[var(--nav-link-active-bg)] text-[var(--color-text-primary)]'
                : 'bg-[var(--color-bg-muted)] text-[var(--color-text-primary)]'
            "
          >
            {{ msg.content }}
            <span v-if="msg.role === 'assistant' && idx === messages.length - 1 && loading">▋</span>
          </div>

          <!-- 引用来源 -->
          <div
            v-if="msg.sources && msg.sources.length > 0"
            class="max-w-[85%] flex flex-col gap-1 mt-1"
          >
            <p class="text-xs text-[var(--color-text-secondary)] opacity-70">
              {{ t('chatPanel.sources') }}
            </p>
            <button
              v-for="src in msg.sources"
              :key="src.echo_id"
              class="text-left text-xs px-2 py-1 rounded border text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)] hover:bg-[var(--color-bg-muted)] transition-colors truncate"
              :style="{ borderColor: 'var(--color-border-subtle)' }"
              @click="goToEcho(src.echo_id)"
            >
              {{ formatSource(src) }}
            </button>
          </div>
        </div>
      </div>

      <!-- 输入区 -->
      <div class="flex flex-row items-end gap-2 mt-4">
        <BaseTextArea
          v-model="input"
          :placeholder="t('chatPanel.inputPlaceholder')"
          class="flex-1"
          :rows="2"
          @keydown="handleKeydown"
        />
        <BaseButton
          :icon="Trumpet"
          :disabled="loading || input.trim().length === 0"
          @click="send(input)"
        >
          {{ t('chatPanel.send') }}
        </BaseButton>
      </div>
    </div>
  </PanelCard>
</template>

<script setup lang="ts">
import PanelCard from '@/layout/PanelCard.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseTextArea from '@/components/common/BaseTextArea.vue'
import Trumpet from '@/components/icons/trumpet.vue'
import Delete from '@/components/icons/delete.vue'
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
let abort: (() => void) | null = null

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

const formatSource = (src: App.Api.Chat.ChatSource): string => {
  const day = new Date(src.echo_created * 1000).toISOString().slice(0, 10)
  const text = src.content.length > 40 ? src.content.slice(0, 40) + '…' : src.content
  return `[${day}] ${text}`
}

const goToEcho = (echoId: string) => {
  router.push(`/echo/${echoId}`)
}

const send = (question: string) => {
  const q = question.trim()
  if (q.length === 0 || loading.value) return

  messages.value.push({ role: 'user', content: q })
  const assistant = ref<App.Api.Chat.ChatMessage>({ role: 'assistant', content: '', sources: [] })
  messages.value.push(assistant.value)
  input.value = ''
  loading.value = true
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

const handleClear = () => {
  if (abort) abort()
  abort = null
  loading.value = false
  messages.value = []
}
</script>

<style scoped></style>
