<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="w-full px-2">
    <!-- OAuth2 设置 -->
    <TheOAuth2Setting class="mb-3" />

    <!-- Passkey 设置 -->
    <ThePasskeySetting class="mb-3" />
  </div>
</template>

<script setup lang="ts">
import TheOAuth2Setting from './TheSetting/TheOAuth2Setting.vue'
import ThePasskeySetting from './TheSetting/ThePasskeySetting.vue'
</script>

<style scoped></style>
