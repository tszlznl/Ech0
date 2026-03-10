<template>
  <article class="markdown-preview-card">
    <header class="preview-card-header">预览</header>
    <div class="preview-card-body">
      <MarkdownRenderer :content="content" />
    </div>
  </article>
</template>

<script setup lang="ts">
import MarkdownRenderer from './MarkdownRenderer.vue'

defineProps<{
  content: string
}>()
</script>
