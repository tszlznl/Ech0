<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="w-full px-2">
    <!-- 对话 -->
    <TheChatBox class="mb-3" />

    <!-- Embedding 向量设置 -->
    <TheEmbeddingSetting class="mb-3" />
  </div>
</template>

<script setup lang="ts">
import TheChatBox from './TheChat/TheChatBox.vue'
import TheEmbeddingSetting from './TheChat/TheEmbeddingSetting.vue'
</script>
