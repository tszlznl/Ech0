<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="px-4 pb-4 py-2 mx-auto flex flex-col w-sm md:w-sm mt-4 mb-12">
    <h1
      class="text-4xl md:text-6xl italic font-bold font-serif text-center text-[var(--color-text-muted)] mb-8"
    >
      Ech0 Widget
    </h1>

    <div class="px-8 mb-2">
      <BaseButton
        @click="router.push('/')"
        class="text-[var(--color-text-primary)] rounded-md shadow-none! border-none! ring-0! bg-transparent! group"
        :tooltip="t('commonNav.backHome')"
      >
        <Arrow
          class="w-9 h-9 rotate-180 transition-transform duration-200 group-hover:-translate-x-1"
        />
      </BaseButton>
    </div>

    <div class="flex flex-col gap-4">
      <TheHeatMap />
      <TheRecentCard v-if="AgentSetting.enable" />
      <TheConnectWidget />
      <TheCommentWidget />
    </div>
  </div>
</template>

<script setup lang="ts">
import BaseButton from '@/components/common/BaseButton.vue'
import Arrow from '@/components/icons/arrow.vue'
import {
  TheCommentWidget,
  TheConnectWidget,
  TheHeatMap,
  TheRecentCard,
} from '@/components/advanced/widget'
import { useSettingStore } from '@/stores'
import { storeToRefs } from 'pinia'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'

const router = useRouter()
const { t } = useI18n()
const settingStore = useSettingStore()
const { AgentSetting } = storeToRefs(settingStore)
</script>
