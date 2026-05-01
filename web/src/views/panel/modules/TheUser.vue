<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="w-full px-2">
    <!-- 用户设置 -->
    <TheUserSetting class="mb-3" />

    <!-- 用户管理 -->
    <TheUserManager class="mb-3" />
  </div>
</template>

<script setup lang="ts">
import TheUserSetting from './TheSetting/TheUserSetting.vue'
import TheUserManager from './TheSetting/TheUserManager.vue'
</script>
