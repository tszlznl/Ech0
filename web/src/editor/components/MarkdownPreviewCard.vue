<template>
  <article class="markdown-preview-card" :class="{ 'is-long-content': isLongContent }">
    <header class="preview-card-header">预览</header>
    <div class="preview-card-body">
      <MarkdownRenderer :content="content" />
    </div>
  </article>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import MarkdownRenderer from './MarkdownRenderer.vue'

const props = defineProps<{
  content: string
}>()

const isLongContent = computed(() => props.content.trim().length >= 260)
</script>
