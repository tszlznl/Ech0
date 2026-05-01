<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <PanelCard>
    <!-- Agent 设置 -->
    <div class="w-full">
      <div class="flex flex-row items-center justify-between mb-3">
        <h1 class="text-[var(--color-text-primary)] font-bold text-lg">
          {{ t('agentSetting.title') }}
        </h1>
        <div class="flex flex-row items-center justify-end">
          <BaseEditCapsule
            :editing="agentEditMode"
            :apply-title="t('commonUi.apply')"
            :cancel-title="t('commonUi.cancel')"
            :edit-title="t('commonUi.edit')"
            @apply="handleUpdateAgentSetting"
            @toggle="agentEditMode = !agentEditMode"
          />
        </div>
      </div>

      <!-- 启用 Agent -->
      <div class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] h-10">
        <h2 class="font-semibold min-w-24 w-max shrink-0 whitespace-nowrap">
          {{ t('agentSetting.enableAgent') }}:
        </h2>
        <BaseSwitch v-model="AgentSetting.enable" :disabled="!agentEditMode" />
      </div>

      <!-- LLM 提供商 -->
      <div
        class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 h-10"
      >
        <h2 class="font-semibold min-w-24 w-max shrink-0 whitespace-nowrap">
          {{ t('agentSetting.provider') }}:
        </h2>
        <BaseSelect
          v-model="AgentSetting.provider"
          :options="agentProviderOptions"
          :disabled="!agentEditMode"
          class="w-40 h-8"
        />
      </div>

      <!-- 模型名称 -->
      <div
        class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 h-10"
      >
        <h2 class="font-semibold min-w-24 w-max shrink-0 whitespace-nowrap">
          {{ t('agentSetting.modelName') }}:
        </h2>
        <span
          v-if="!agentEditMode"
          class="flex-1 min-w-0 truncate inline-block align-middle"
          v-tooltip="AgentSetting.model"
        >
          {{ AgentSetting.model || t('commonUi.none') }}
        </span>
        <div v-else>
          <BaseInput
            v-model="AgentSetting.model"
            type="text"
            :placeholder="t('agentSetting.modelPlaceholder')"
            class="w-full py-1!"
          />
        </div>
      </div>

      <!-- API Key -->
      <div
        class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 h-10"
      >
        <h2 class="font-semibold min-w-24 w-max shrink-0 whitespace-nowrap">
          {{ t('agentSetting.apiKey') }}:
        </h2>
        <span v-if="!agentEditMode" class="flex-1 min-w-0 truncate inline-block align-middle">
          {{ AgentSetting.api_key ? '********' : t('commonUi.none') }}
        </span>
        <BaseInput
          v-else
          v-model="AgentSetting.api_key"
          type="password"
          :placeholder="t('agentSetting.apiKeyPlaceholder')"
          class="w-full py-1!"
        />
      </div>

      <!-- 自定义 Base URL -->
      <div class="flex justify-start text-[var(--color-text-secondary)] gap-2 mt-2">
        <h2 class="font-semibold min-w-24 w-max shrink-0 whitespace-nowrap">
          {{ t('agentSetting.baseUrl') }}:
        </h2>
        <span v-if="!agentEditMode" class="flex-1 min-w-0 truncate inline-block align-middle">
          {{ AgentSetting.base_url.length == 0 ? t('commonUi.none') : AgentSetting.base_url }}
        </span>
        <BaseInput
          v-if="agentEditMode"
          v-model="AgentSetting.base_url"
          :placeholder="t('agentSetting.baseUrlPlaceholder')"
          class="w-full py-1!"
        />
      </div>

      <!-- Prompt -->
      <div class="flex justify-start text-[var(--color-text-secondary)] gap-2 mt-2">
        <h2 class="font-semibold min-w-24 w-max shrink-0 whitespace-nowrap">
          {{ t('agentSetting.prompt') }}:
        </h2>
        <span v-if="!agentEditMode" class="flex-1 min-w-0 truncate inline-block align-middle">
          {{ AgentSetting.prompt.length == 0 ? t('commonUi.none') : '' }}
        </span>
        <BaseTextArea
          v-if="agentEditMode"
          v-model="AgentSetting.prompt"
          :placeholder="t('agentSetting.promptPlaceholder')"
          class="w-full"
          :rows="4"
        />
      </div>
    </div>
  </PanelCard>
</template>

<script setup lang="ts">
import PanelCard from '@/layout/PanelCard.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseSwitch from '@/components/common/BaseSwitch.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import BaseTextArea from '@/components/common/BaseTextArea.vue'
import BaseEditCapsule from '@/components/common/BaseEditCapsule.vue'
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { fetchUpdateAgentSettings } from '@/service/api'
import { theToast } from '@/utils/toast'
import { useSettingStore } from '@/stores'
import { storeToRefs } from 'pinia'
import { AgentProvider } from '@/enums/enums'

const settingStore = useSettingStore()
const { t } = useI18n()
const { getAgentSetting } = settingStore
const { AgentSetting } = storeToRefs(settingStore)

const agentEditMode = ref<boolean>(false)

const agentProviderOptions = ref<{ label: string; value: AgentProvider }[]>([
  { label: 'OpenAI', value: AgentProvider.OPENAI },
  { label: 'Anthropic', value: AgentProvider.ANTHROPIC },
  { label: 'Gemini', value: AgentProvider.GEMINI },
])

const handleUpdateAgentSetting = async () => {
  await fetchUpdateAgentSettings(settingStore.AgentSetting)
    .then((res) => {
      if (res.code === 1) {
        theToast.success(res.msg)
      }
    })
    .finally(() => {
      agentEditMode.value = false
      // 重新获取 Agent 设置
      getAgentSetting()
    })
}

onMounted(() => {
  getAgentSetting()
})
</script>

<style scoped></style>
