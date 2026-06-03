<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="w-full px-2">
    <!-- 分段控件：OAuth2 / Passkey -->
    <BaseSegmented v-model="tab" :options="tabOptions" />

    <!-- OAuth2（设置 + 账号绑定） -->
    <TheOAuth2Setting v-if="tab === 'oauth2'" />
    <!-- Passkey -->
    <ThePasskeySetting v-else />
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseSegmented from '@/components/common/BaseSegmented.vue'
import TheOAuth2Setting from './TheSetting/TheOAuth2Setting.vue'
import ThePasskeySetting from './TheSetting/ThePasskeySetting.vue'

const { t } = useI18n()
const tab = ref('oauth2')
const tabOptions = computed(() => [
  { label: String(t('ssoManagement.tabOAuth2')), value: 'oauth2' },
  { label: String(t('ssoManagement.tabPasskey')), value: 'passkey' },
])
</script>

<style scoped></style>
