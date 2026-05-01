<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <section class="echo-interactions w-full max-w-sm mx-auto">
    <TheComment />
  </section>
</template>

<script setup lang="ts">
import TheComment from '@/components/advanced/TheComment.vue'
</script>

<style scoped>
.echo-interactions {
  border-top: 1px dashed var(--color-border-subtle);
}
</style>
