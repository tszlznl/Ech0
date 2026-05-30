<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="chat-page">
    <div class="chat-container">
      <div class="chat-topbar">
        <BaseButton
          :icon="Back"
          @click="router.push('/')"
          :tooltip="t('commonNav.backHome')"
          class="chat-back-btn"
        />
      </div>
      <TheChatBox />
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import BaseButton from '@/components/common/BaseButton.vue'
import Back from '@/components/icons/back.vue'
import TheChatBox from './TheChatBox.vue'

const { t } = useI18n()
const router = useRouter()
</script>

<style scoped>
.chat-page {
  min-height: calc(100vh - 3.5rem);
  display: flex;
  justify-content: center;
  padding: 1rem 0.75rem 2rem;
}

.chat-container {
  width: 100%;
  max-width: 48rem;
}

.chat-topbar {
  display: flex;
  justify-content: flex-start;
  margin-bottom: 0.75rem;
}

.chat-back-btn {
  border: 1px solid var(--color-border-subtle) !important;
  border-radius: 999px !important;
  background: var(--color-bg-surface) !important;
  color: var(--color-text-secondary) !important;
  width: 2rem !important;
  height: 2rem !important;
  padding: 0 !important;
  display: inline-flex !important;
  align-items: center !important;
  justify-content: center !important;
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease;
}

.chat-back-btn:hover {
  border-color: var(--color-border-strong) !important;
  background: var(--color-bg-muted) !important;
}
</style>
