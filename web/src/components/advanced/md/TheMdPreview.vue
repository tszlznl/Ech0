<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <MarkdownRenderer :content="props.content" />
</template>
<script setup lang="ts">
import { MarkdownRenderer } from '@/editor/preview'

const props = defineProps<{
  content: string
}>()
</script>
