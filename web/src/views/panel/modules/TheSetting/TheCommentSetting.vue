<template>
  <PanelCard>
    <div class="w-full">
      <div class="flex flex-row items-center justify-between mb-3">
        <h1 class="text-[var(--color-text-primary)] font-bold text-lg">评论设置</h1>
        <div class="flex flex-row items-center justify-end">
          <BaseEditCapsule
            :editing="commentEditMode"
            apply-title="应用"
            cancel-title="取消"
            edit-title="编辑"
            @apply="handleUpdateCommentSetting"
            @toggle="toggleEdit"
          />
        </div>
      </div>

      <div class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] h-10">
        <h2 class="font-semibold w-24 shrink-0">启用评论:</h2>
        <BaseSwitch v-model="CommentSetting.enable_comment" :disabled="!commentEditMode" />
      </div>

      <div
        class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 h-10"
      >
        <h2 class="font-semibold w-24 shrink-0">评论服务:</h2>
        <BaseSelect
          v-model="CommentSetting.provider"
          :options="commentServiceOptions"
          :disabled="!commentEditMode"
          class="w-34 h-8"
        />
      </div>

      <div
        v-if="currentProviderMeta"
        class="mt-2 rounded-[var(--radius-md)] border border-[var(--color-border-subtle)] p-3"
      >
        <h3 class="text-sm font-semibold text-[var(--color-text-primary)] mb-2">供应商参数</h3>

        <div
          v-for="field in currentProviderMeta.fields"
          :key="field.key"
          class="flex items-center gap-2 text-[var(--color-text-secondary)] min-h-10 mb-2"
        >
          <h4 class="font-semibold w-24 shrink-0">
            {{ field.label }}<span v-if="field.required" class="text-red-400">*</span>:
          </h4>

          <template v-if="!commentEditMode">
            <span class="truncate max-w-60" :title="displayConfigValue(field.key)">
              {{ displayConfigValue(field.key) || '暂无' }}
            </span>
          </template>

          <template v-else>
            <BaseInput
              :model-value="displayConfigValue(field.key)"
              type="text"
              :placeholder="field.placeholder || ''"
              class="w-full py-1!"
              @update:model-value="setConfigValue(field.key, $event)"
            />
          </template>
        </div>
      </div>
    </div>
  </PanelCard>
</template>

<script setup lang="ts">
import PanelCard from '@/layout/PanelCard.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseSwitch from '@/components/common/BaseSwitch.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import BaseEditCapsule from '@/components/common/BaseEditCapsule.vue'
import { ref, onMounted, computed, watch } from 'vue'
import { fetchUpdateCommentSettings } from '@/service/api'
import { theToast } from '@/utils/toast'
import { useSettingStore } from '@/stores'
import { storeToRefs } from 'pinia'

const settingStore = useSettingStore()
const { getCommentSetting, getCommentProviderMeta } = settingStore
const { CommentSetting, CommentProviderMeta } = storeToRefs(settingStore)

const commentEditMode = ref<boolean>(false)

const commentServiceOptions = computed(() =>
  CommentProviderMeta.value.map((item) => ({
    label: item.label,
    value: item.provider,
  })),
)

const currentProviderMeta = computed(() =>
  CommentProviderMeta.value.find((item) => item.provider === CommentSetting.value.provider),
)

const ensureActiveProviderSetting = () => {
  const provider = CommentSetting.value.provider
  if (!CommentSetting.value.providers) {
    CommentSetting.value.providers = {}
  }
  if (!CommentSetting.value.providers[provider]) {
    CommentSetting.value.providers[provider] = { config: {} }
  }
  if (!CommentSetting.value.providers[provider].config) {
    CommentSetting.value.providers[provider].config = {}
  }
}

const displayConfigValue = (key: string) => {
  ensureActiveProviderSetting()
  const provider = CommentSetting.value.provider
  const value = CommentSetting.value.providers[provider].config[key]
  if (value === undefined || value === null) return ''
  return String(value)
}

const setConfigValue = (key: string, value: string | number | null | undefined) => {
  ensureActiveProviderSetting()
  const provider = CommentSetting.value.provider
  CommentSetting.value.providers[provider].config[key] = value ?? ''
}

const handleUpdateCommentSetting = async () => {
  ensureActiveProviderSetting()

  await fetchUpdateCommentSettings(settingStore.CommentSetting)
    .then((res) => {
      if (res.code === 1) {
        theToast.success(res.msg)
      }
    })
    .finally(() => {
      commentEditMode.value = false
      getCommentSetting()
    })
}

const toggleEdit = async () => {
  if (commentEditMode.value) {
    await getCommentSetting()
  }
  commentEditMode.value = !commentEditMode.value
}

watch(
  () => CommentSetting.value.provider,
  () => {
    ensureActiveProviderSetting()
  },
)

onMounted(() => {
  getCommentProviderMeta()
  getCommentSetting()
  ensureActiveProviderSetting()
})
</script>
