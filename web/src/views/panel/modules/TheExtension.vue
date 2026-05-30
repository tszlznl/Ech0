<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="w-full px-2">
    <!-- 连接设置 -->
    <TheConnectSetting class="mb-3" />
    <!-- Ech0 Copilot 设置 -->
    <TheCopilotSetting class="mb-3" />
  </div>
</template>

<script setup lang="ts">
import TheConnectSetting from './TheSetting/TheConnectSetting.vue'
import TheCopilotSetting from './TheCopilot/TheCopilotSetting.vue'
</script>

<style scoped></style>
