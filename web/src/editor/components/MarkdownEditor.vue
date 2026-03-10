<template>
  <div class="markdown-editor">
    <div class="toolbar">
      <button
        v-for="item in toolbarItems"
        :key="item.action"
        type="button"
        class="toolbar-btn"
        @click="onToolbarClick(item.action)"
      >
        {{ item.label }}
      </button>
      <button type="button" class="toolbar-btn ml-auto" @click="showPreview = !showPreview">
        {{ showPreview ? '隐藏预览' : '显示预览' }}
      </button>
    </div>

    <div class="editor-layout" :class="{ 'preview-hidden': !showPreview }">
      <textarea
        ref="textareaRef"
        class="editor-input"
        :placeholder="placeholder"
        :value="modelValue"
        @input="onInput"
      />
      <div v-if="showPreview" class="preview-pane">
        <MarkdownRenderer :content="modelValue" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import MarkdownRenderer from './MarkdownRenderer.vue'
import { applyMarkdownAction } from '../composables/useMarkdownEditorActions'
import type { MarkdownEditorAction } from '../types'

defineProps<{
  modelValue: string
  placeholder?: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

const textareaRef = ref<HTMLTextAreaElement | null>(null)
const showPreview = ref(true)

const toolbarItems: Array<{ label: string; action: MarkdownEditorAction }> = [
  { label: '粗体', action: 'bold' },
  { label: '斜体', action: 'italic' },
  { label: '标题', action: 'heading' },
  { label: '引用', action: 'quote' },
  { label: '无序列表', action: 'unorderedList' },
  { label: '有序列表', action: 'orderedList' },
  { label: '代码块', action: 'codeBlock' },
  { label: '链接', action: 'link' },
]

function onInput(event: Event) {
  const target = event.target as HTMLTextAreaElement
  emit('update:modelValue', target.value)
}

function onToolbarClick(action: MarkdownEditorAction) {
  if (!textareaRef.value) return
  applyMarkdownAction(textareaRef.value, action)
}
</script>
