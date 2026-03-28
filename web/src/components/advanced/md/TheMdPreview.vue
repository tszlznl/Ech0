<template>
  <MarkdownRenderer :content="props.content" />
</template>
<script setup lang="ts">
import { MarkdownRenderer } from '@/editor/preview'

const props = defineProps<{
  content: string
}>()
</script>
