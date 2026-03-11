<template>
  <div class="markdown-editor-root">
    <div v-if="isFullMode" class="markdown-editor-backdrop" @click="exitFullMode"></div>

    <div v-if="!isFullMode" class="markdown-editor">
      <button
        type="button"
        class="mode-toggle-btn"
        title="全屏编辑"
        aria-label="全屏编辑"
        @click="enterFullMode"
      >
        <Full class="w-3.5 h-3.5" />
      </button>
      <div class="editor-content">
        <textarea
          ref="textareaRef"
          class="editor-input"
          :placeholder="placeholder"
          :value="modelValue"
          @input="onInput"
        />
      </div>
    </div>

    <div
      v-else
      class="markdown-full-shell"
      :class="{ 'no-preview': isPreviewMode }"
    >
      <div class="markdown-editor is-full">
        <div class="toolbar">
        <button
          v-for="item in toolbarItems"
          :key="item.action"
          type="button"
          class="toolbar-btn"
          :title="item.label"
          :aria-label="item.label"
          @click="onToolbarClick(item.action)"
        >
          <span class="toolbar-icon" aria-hidden="true">{{ item.icon }}</span>
        </button>
        <div class="toolbar-actions">
          <button
            type="button"
            class="toolbar-btn"
            :title="isPreviewMode ? '显示预览' : '隐藏预览'"
            :aria-label="isPreviewMode ? '显示预览' : '隐藏预览'"
            @click="isPreviewMode = !isPreviewMode"
          >
            <Preview class="w-3.5 h-3.5" />
          </button>
          <button
            type="button"
            class="toolbar-btn mode-toggle-btn-inline"
            title="退出全屏"
            aria-label="退出全屏"
            @click="exitFullMode"
          >
            <Closefull class="w-3.5 h-3.5" />
          </button>
        </div>
        </div>

        <div class="editor-content">
        <textarea
          ref="textareaRef"
          class="editor-input"
          :placeholder="placeholder"
          :value="modelValue"
          @input="onInput"
        />
        </div>
      </div>

      <div v-if="!isPreviewMode" class="markdown-preview-dock">
        <MarkdownPreviewCard :content="modelValue" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import MarkdownPreviewCard from './MarkdownPreviewCard.vue'
import { applyMarkdownAction } from '../composables/useMarkdownEditorActions'
import type { MarkdownEditorAction } from '../types'
import Full from '@/components/icons/full.vue'
import Closefull from '@/components/icons/closefull.vue'
import Preview from '@/components/icons/preview.vue'

defineProps<{
  modelValue: string
  placeholder?: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

const textareaRef = ref<HTMLTextAreaElement | null>(null)
const isFullMode = ref(false)
const isPreviewMode = ref(false)

const toolbarItems: Array<{ label: string; icon: string; action: MarkdownEditorAction }> = [
  { label: '粗体', icon: 'B', action: 'bold' },
  { label: '斜体', icon: 'I', action: 'italic' },
  { label: '标题', icon: 'H', action: 'heading' },
  { label: '引用', icon: '❝', action: 'quote' },
  { label: '无序列表', icon: '•', action: 'unorderedList' },
]

function onInput(event: Event) {
  const target = event.target as HTMLTextAreaElement
  emit('update:modelValue', target.value)
}

function onToolbarClick(action: MarkdownEditorAction) {
  if (isPreviewMode.value) {
    isPreviewMode.value = false
    nextTick(() => {
      if (!textareaRef.value) return
      applyMarkdownAction(textareaRef.value, action)
    })
    return
  }

  if (!textareaRef.value) return
  applyMarkdownAction(textareaRef.value, action)
}

function enterFullMode() {
  isFullMode.value = true
  isPreviewMode.value = false
}

function exitFullMode() {
  isFullMode.value = false
  isPreviewMode.value = false
}

function onWindowKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape' && isFullMode.value) {
    exitFullMode()
  }
}

onMounted(() => {
  window.addEventListener('keydown', onWindowKeydown)
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', onWindowKeydown)
})
</script>
