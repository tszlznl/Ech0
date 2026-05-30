<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <!-- 模型 -->
  <div class="w-full">
    <!-- 启用 Agent -->
    <div class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] h-10">
      <h2 class="font-semibold min-w-24 w-max shrink-0 whitespace-nowrap">
        {{ t('agentSetting.enableAgent') }}:
      </h2>
      <BaseSwitch v-model="AgentSetting.enable" :disabled="!editMode" />
    </div>

    <!-- LLM 接口协议 -->
    <div
      class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 h-10"
    >
      <h2 class="font-semibold min-w-24 w-max shrink-0 whitespace-nowrap">
        {{ t('agentSetting.protocol') }}:
      </h2>
      <BaseSelect
        v-model="AgentSetting.protocol"
        :options="agentProtocolOptions"
        :disabled="!editMode"
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
        v-if="!editMode"
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
      <span v-if="!editMode" class="flex-1 min-w-0 truncate inline-block align-middle">
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
      <span v-if="!editMode" class="flex-1 min-w-0 truncate inline-block align-middle">
        {{ AgentSetting.base_url.length == 0 ? t('commonUi.none') : AgentSetting.base_url }}
      </span>
      <BaseInput
        v-if="editMode"
        v-model="AgentSetting.base_url"
        :placeholder="t('agentSetting.baseUrlPlaceholder')"
        class="w-full py-1!"
      />
    </div>
    <p v-if="editMode" class="text-xs text-[var(--color-text-secondary)] opacity-70 mt-1 ml-26">
      {{ t('agentSetting.baseUrlHint') }}
    </p>

    <!-- Prompt -->
    <div class="flex justify-start text-[var(--color-text-secondary)] gap-2 mt-2">
      <h2 class="font-semibold min-w-24 w-max shrink-0 whitespace-nowrap">
        {{ t('agentSetting.prompt') }}:
      </h2>
      <span v-if="!editMode" class="flex-1 min-w-0 truncate inline-block align-middle">
        {{ AgentSetting.prompt.length == 0 ? t('commonUi.none') : '' }}
      </span>
      <BaseTextArea
        v-if="editMode"
        v-model="AgentSetting.prompt"
        :placeholder="t('agentSetting.promptPlaceholder')"
        class="w-full"
        :rows="4"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import BaseInput from '@/components/common/BaseInput.vue'
import BaseSwitch from '@/components/common/BaseSwitch.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import BaseTextArea from '@/components/common/BaseTextArea.vue'
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { fetchUpdateAgentSettings } from '@/service/api'
import { theToast } from '@/utils/toast'
import { useSettingStore } from '@/stores'
import { storeToRefs } from 'pinia'
import { AgentProtocol } from '@/enums/enums'

defineProps<{ editMode: boolean }>()

const settingStore = useSettingStore()
const { t } = useI18n()
const { getAgentSetting } = settingStore
const { AgentSetting } = storeToRefs(settingStore)

const agentProtocolOptions = computed<{ label: string; value: AgentProtocol }[]>(() => [
  { label: t('agentSetting.protocolOpenAI'), value: AgentProtocol.OPENAI },
  { label: t('agentSetting.protocolAnthropic'), value: AgentProtocol.ANTHROPIC },
  { label: t('agentSetting.protocolGemini'), value: AgentProtocol.GEMINI },
])

// 由父组件的编辑胶囊触发；保存后回填最新设置
const save = async () => {
  await fetchUpdateAgentSettings(settingStore.AgentSetting)
    .then((res) => {
      if (res.code === 1) {
        theToast.success(res.msg)
      }
    })
    .finally(() => {
      getAgentSetting()
    })
}

defineExpose({ save })

onMounted(() => {
  getAgentSetting()
})
</script>

<style scoped></style>
