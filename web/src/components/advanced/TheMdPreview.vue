<template>
  <MarkdownRenderer :content="props.content" />
</template>
<script setup lang="ts">
import { MarkdownRenderer } from '@/editor'

const props = defineProps<{
  content: string
}>()
</script>
