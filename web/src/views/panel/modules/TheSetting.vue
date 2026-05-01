<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="w-full px-2">
    <!-- 系统设置 -->
    <TheSystemSetting class="mb-3" />

    <!-- 访问令牌设置 -->
    <TheAccessTokenSetting class="mb-3" />
  </div>
</template>

<script setup lang="ts">
import TheSystemSetting from './TheSetting/TheSystemSetting.vue'
import TheAccessTokenSetting from './TheSetting/TheAccessTokenSetting.vue'
</script>
