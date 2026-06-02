<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="w-full text-[var(--color-text-secondary)]">
    <!-- 可选说明：向量索引非必需，不配置也能正常对话 -->
    <p class="text-xs opacity-70 mb-4">{{ t('embeddingSetting.optionalHint') }}</p>

    <!-- 启用 -->
    <div class="flex items-center justify-between mb-4">
      <h2 class="font-semibold">{{ t('embeddingSetting.enable') }}</h2>
      <BaseSwitch v-model="setting.enable" :disabled="!editMode" />
    </div>

    <!-- 模型名称 -->
    <div class="mb-4">
      <h2 class="font-semibold mb-1.5">{{ t('embeddingSetting.modelName') }}</h2>
      <span v-if="!editMode" class="block truncate opacity-80" v-tooltip="setting.model">
        {{ setting.model || t('commonUi.none') }}
      </span>
      <BaseCombobox
        v-else
        v-model="setting.model"
        :options="modelOptions"
        :allow-create="true"
        :placeholder="t('embeddingSetting.modelPlaceholder')"
        wrapper-class="w-full"
        class="w-full"
      />
    </div>

    <!-- 向量维度 -->
    <div class="mb-4">
      <h2 class="font-semibold mb-1.5">{{ t('embeddingSetting.dim') }}</h2>
      <span v-if="!editMode" class="block truncate opacity-80">
        {{ setting.dim || t('commonUi.none') }}
      </span>
      <BaseInput
        v-else
        v-model.number="setting.dim"
        type="number"
        :placeholder="t('embeddingSetting.dimPlaceholder')"
        class="w-full"
      />
    </div>

    <!-- API Key -->
    <div class="mb-4">
      <h2 class="font-semibold mb-1.5">{{ t('embeddingSetting.apiKey') }}</h2>
      <span v-if="!editMode" class="block truncate opacity-80">
        {{ setting.api_key ? '********' : t('commonUi.none') }}
      </span>
      <BaseInput
        v-else
        v-model="setting.api_key"
        type="password"
        :placeholder="t('embeddingSetting.apiKeyPlaceholder')"
        class="w-full"
      />
    </div>

    <!-- 自定义 Base URL -->
    <div class="mb-4">
      <h2 class="font-semibold mb-1.5">{{ t('embeddingSetting.baseUrl') }}</h2>
      <span v-if="!editMode" class="block truncate opacity-80">
        {{ setting.base_url.length === 0 ? t('commonUi.none') : setting.base_url }}
      </span>
      <BaseInput
        v-else
        v-model="setting.base_url"
        :placeholder="t('embeddingSetting.baseUrlPlaceholder')"
        class="w-full"
      />
      <p v-if="editMode" class="text-xs opacity-70 mt-1">{{ t('embeddingSetting.baseUrlHint') }}</p>
    </div>

    <!-- 重建索引（异步作业 + 轮询进度，可取消） -->
    <div
      class="flex flex-row items-center justify-between gap-2 mt-5 pt-4 border-t border-[var(--color-border-subtle)]"
    >
      <div class="min-w-0">
        <h2 class="font-semibold">{{ t('embeddingSetting.reindex') }}</h2>
        <p class="text-xs opacity-70 mt-1">{{ t('embeddingSetting.reindexHint') }}</p>
      </div>
      <div class="flex items-center gap-2 shrink-0">
        <BaseButton v-if="reindex.isRunning" @click="handleCancelReindex">
          {{ t('embeddingSetting.reindexCancel') }}
        </BaseButton>
        <BaseButton
          :loading="reindex.isRunning"
          :disabled="reindex.isRunning"
          @click="handleReindex"
        >
          {{ t('embeddingSetting.reindexAction') }}
        </BaseButton>
      </div>
    </div>
    <!-- 进行中：有计数则显示已索引进度，否则显示通用进行中 -->
    <p v-if="reindex.isRunning" class="text-xs text-[var(--color-text-secondary)] opacity-80 mt-2">
      {{
        reindex.result
          ? t('embeddingSetting.reindexProgress', {
              indexed: reindex.result.indexed,
              total: reindex.result.total,
            })
          : t('embeddingSetting.reindexRunning')
      }}
    </p>
    <!-- 成功 -->
    <p
      v-else-if="reindex.status === 'success' && reindex.result"
      class="text-xs text-[var(--color-text-secondary)] opacity-80 mt-2"
    >
      {{
        t('embeddingSetting.reindexResult', {
          indexed: reindex.result.indexed,
          total: reindex.result.total,
          failed: reindex.result.failed,
        })
      }}
    </p>
    <!-- 失败 -->
    <p
      v-else-if="reindex.status === 'failed'"
      class="text-xs text-[var(--color-danger,#dc2626)] mt-2"
    >
      {{ t('embeddingSetting.reindexFailed', { error: reindex.error }) }}
    </p>
    <!-- 已取消 -->
    <p
      v-else-if="reindex.status === 'cancelled'"
      class="text-xs text-[var(--color-text-secondary)] opacity-80 mt-2"
    >
      {{ t('embeddingSetting.reindexCancelled') }}
    </p>
  </div>
</template>

<script setup lang="ts">
import BaseInput from '@/components/common/BaseInput.vue'
import BaseSwitch from '@/components/common/BaseSwitch.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseCombobox from '@/components/common/BaseCombobox.vue'
import { ref, watch, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { fetchGetEmbeddingSettings, fetchUpdateEmbeddingSettings } from '@/service/api'
import { theToast } from '@/utils/toast'
import { useBaseDialog } from '@/composables/useBaseDialog'
import { useReindexStore } from '@/stores/reindex'

const props = defineProps<{ editMode: boolean }>()

const { t } = useI18n()
const { openConfirm } = useBaseDialog()

// 常见 OpenAI 兼容 embedding 模型 → 默认维度（提示性，非穷举；维度仍可手动改）
const MODEL_DIM_PRESETS: Record<string, number> = {
  'text-embedding-3-small': 1536,
  'text-embedding-3-large': 3072,
  'text-embedding-ada-002': 1536,
  'text-embedding-v3': 1024, // Qwen / DashScope
  'bge-m3': 1024,
  'bge-large-zh-v1.5': 1024,
  'jina-embeddings-v3': 1024,
  'nomic-embed-text': 768, // Ollama
  'mxbai-embed-large': 1024, // Ollama
}
const modelOptions = Object.keys(MODEL_DIM_PRESETS)

// 重建索引改为异步作业，状态/进度/取消全交给 reindex store（复用 migration 轮询范式）。
const reindex = useReindexStore()

const setting = ref<App.Api.Embedding.EmbeddingSetting>({
  enable: false,
  model: '',
  api_key: '',
  base_url: '',
  dim: 0,
})

// 已保存的基线，用于判断 model/dim 是否变化（变化则需重建索引）
const originalModel = ref<string>('')
const originalDim = ref<number>(0)

// 命中预设的模型时自动带出其默认维度。需要自定义维度（如 Matryoshka 模型）时，
// 选完模型再手动改维度即可——之后模型未变，本 watch 不会再覆盖手填值。
watch(
  () => setting.value.model,
  (next) => {
    if (!props.editMode) return
    const preset = MODEL_DIM_PRESETS[next]
    if (preset) {
      setting.value.dim = preset
    }
  },
)

const getSetting = async () => {
  const res = await fetchGetEmbeddingSettings()
  if (res.code === 1 && res.data) {
    setting.value = res.data
    originalModel.value = res.data.model
    originalDim.value = res.data.dim
  }
}

// 由父组件的编辑胶囊触发；保存后回填最新设置
const save = async () => {
  const changed =
    setting.value.model !== originalModel.value || setting.value.dim !== originalDim.value
  await fetchUpdateEmbeddingSettings(setting.value)
    .then((res) => {
      if (res.code === 1) {
        theToast.success(res.msg)
        // 模型/维度变更且索引已启用：旧索引已失效，提示用户立即重建
        if (changed && setting.value.enable) {
          openConfirm({
            title: t('embeddingSetting.reindexConfirmTitle'),
            description: t('embeddingSetting.reindexConfirmDesc'),
            onConfirm: () => handleReindex(),
          })
        }
      }
    })
    .finally(() => {
      getSetting()
    })
}

defineExpose({ save })

const handleReindex = async () => {
  const res = await reindex.start()
  if (res.code === 1) {
    theToast.success(res.msg)
  }
}

const handleCancelReindex = async () => {
  const res = await reindex.cancel()
  if (res.code === 1) {
    theToast.success(res.msg)
  }
}

onMounted(() => {
  getSetting()
  reindex.init()
})

onUnmounted(() => {
  reindex.stopPolling()
})
</script>

<style scoped></style>
