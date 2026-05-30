<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="w-full">
    <p class="text-xs text-[var(--color-text-secondary)] opacity-70 mb-3">
      {{ t('embeddingSetting.intro') }}
    </p>

    <!-- 启用 -->
    <div class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] h-10">
      <h2 class="font-semibold min-w-24 w-max shrink-0 whitespace-nowrap">
        {{ t('embeddingSetting.enable') }}:
      </h2>
      <BaseSwitch v-model="setting.enable" :disabled="!editMode" />
    </div>

    <!-- 模型名称 -->
    <div
      class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 h-10"
    >
      <h2 class="font-semibold min-w-24 w-max shrink-0 whitespace-nowrap">
        {{ t('embeddingSetting.modelName') }}:
      </h2>
      <span
        v-if="!editMode"
        class="flex-1 min-w-0 truncate inline-block align-middle"
        v-tooltip="setting.model"
      >
        {{ setting.model || t('commonUi.none') }}
      </span>
      <BaseInput
        v-else
        v-model="setting.model"
        type="text"
        :placeholder="t('embeddingSetting.modelPlaceholder')"
        class="w-full py-1!"
      />
    </div>

    <!-- 向量维度 -->
    <div
      class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 h-10"
    >
      <h2 class="font-semibold min-w-24 w-max shrink-0 whitespace-nowrap">
        {{ t('embeddingSetting.dim') }}:
      </h2>
      <span v-if="!editMode" class="flex-1 min-w-0 truncate inline-block align-middle">
        {{ setting.dim || t('commonUi.none') }}
      </span>
      <BaseInput
        v-else
        v-model.number="setting.dim"
        type="number"
        :placeholder="t('embeddingSetting.dimPlaceholder')"
        class="w-full py-1!"
      />
    </div>

    <!-- API Key -->
    <div
      class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 h-10"
    >
      <h2 class="font-semibold min-w-24 w-max shrink-0 whitespace-nowrap">
        {{ t('embeddingSetting.apiKey') }}:
      </h2>
      <span v-if="!editMode" class="flex-1 min-w-0 truncate inline-block align-middle">
        {{ setting.api_key ? '********' : t('commonUi.none') }}
      </span>
      <BaseInput
        v-else
        v-model="setting.api_key"
        type="password"
        :placeholder="t('embeddingSetting.apiKeyPlaceholder')"
        class="w-full py-1!"
      />
    </div>

    <!-- 自定义 Base URL -->
    <div class="flex justify-start text-[var(--color-text-secondary)] gap-2 mt-2">
      <h2 class="font-semibold min-w-24 w-max shrink-0 whitespace-nowrap">
        {{ t('embeddingSetting.baseUrl') }}:
      </h2>
      <span v-if="!editMode" class="flex-1 min-w-0 truncate inline-block align-middle">
        {{ setting.base_url.length == 0 ? t('commonUi.none') : setting.base_url }}
      </span>
      <BaseInput
        v-if="editMode"
        v-model="setting.base_url"
        :placeholder="t('embeddingSetting.baseUrlPlaceholder')"
        class="w-full py-1!"
      />
    </div>
    <p v-if="editMode" class="text-xs text-[var(--color-text-secondary)] opacity-70 mt-1 ml-26">
      {{ t('embeddingSetting.baseUrlHint') }}
    </p>

    <!-- 重建索引 -->
    <div class="flex flex-row items-center justify-between mt-4">
      <div class="text-[var(--color-text-secondary)]">
        <h2 class="font-semibold">{{ t('embeddingSetting.reindex') }}</h2>
        <p class="text-xs opacity-70 mt-1">{{ t('embeddingSetting.reindexHint') }}</p>
      </div>
      <BaseButton :loading="reindexing" :disabled="reindexing" @click="handleReindex">
        {{ t('embeddingSetting.reindexAction') }}
      </BaseButton>
    </div>
    <p v-if="reindexResult" class="text-xs text-[var(--color-text-secondary)] opacity-80 mt-2">
      {{
        t('embeddingSetting.reindexResult', {
          indexed: reindexResult.indexed,
          total: reindexResult.total,
          failed: reindexResult.failed,
        })
      }}
    </p>
  </div>
</template>

<script setup lang="ts">
import BaseInput from '@/components/common/BaseInput.vue'
import BaseSwitch from '@/components/common/BaseSwitch.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  fetchGetEmbeddingSettings,
  fetchUpdateEmbeddingSettings,
  fetchReindexEmbeddings,
} from '@/service/api'
import { theToast } from '@/utils/toast'

defineProps<{ editMode: boolean }>()

const { t } = useI18n()

const reindexing = ref<boolean>(false)
const reindexResult = ref<App.Api.Embedding.ReindexResult | null>(null)

const setting = ref<App.Api.Embedding.EmbeddingSetting>({
  enable: false,
  model: '',
  api_key: '',
  base_url: '',
  dim: 0,
})

const getSetting = async () => {
  const res = await fetchGetEmbeddingSettings()
  if (res.code === 1 && res.data) {
    setting.value = res.data
  }
}

// 由父组件的编辑胶囊触发；保存后回填最新设置
const save = async () => {
  await fetchUpdateEmbeddingSettings(setting.value)
    .then((res) => {
      if (res.code === 1) {
        theToast.success(res.msg)
      }
    })
    .finally(() => {
      getSetting()
    })
}

defineExpose({ save })

const handleReindex = async () => {
  reindexing.value = true
  reindexResult.value = null
  try {
    const res = await fetchReindexEmbeddings()
    if (res.code === 1 && res.data) {
      reindexResult.value = res.data
      theToast.success(res.msg)
    }
  } finally {
    reindexing.value = false
  }
}

onMounted(() => {
  getSetting()
})
</script>

<style scoped></style>
