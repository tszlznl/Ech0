<template>
  <PanelCard>
    <!-- Webhook 设置 -->
    <div class="w-full">
      <div class="flex flex-row items-center justify-between mb-4">
        <h1 class="text-[var(--color-text-primary)] font-bold text-lg">Webhook</h1>
        <div class="flex flex-row items-center justify-end">
          <BaseEditCapsule
            :editing="webhookEdit"
            apply-title="完成"
            cancel-title="取消"
            edit-title="编辑"
            @apply="webhookEdit = false"
            @toggle="webhookEdit = !webhookEdit"
          />
        </div>
      </div>

      <!-- 添加 Webhook -->
      <div
        v-if="webhookEdit"
        class="mb-2 border border-[var(--color-border-subtle)] border-dashed rounded-[var(--radius-md)] flex flex-col gap-2 p-2 text-[var(--color-text-muted)]"
      >
        <div>
          <span>Webhook 名称：</span>
          <BaseInput class="w-full" v-model="webhookToAdd.name" placeholder="Webhook 名称" />
        </div>

        <div>
          <span>Webhook 地址：</span>
          <BaseInput
            class="w-full"
            v-model="webhookToAdd.url"
            placeholder="Webhook 地址（带https/http）"
          />
        </div>

        <div class="flex items-center justify-center my-2">
          <BaseButton
            :disabled="isSubmitting"
            @click="handleCancelAddWebhook"
            class="w-1/3 h-8 rounded-md flex justify-center mr-2"
            title="取消添加"
          >
            <span>取消</span>
          </BaseButton>

          <BaseButton
            :loading="isSubmitting"
            @click="handleAddWebhook"
            class="w-1/3 h-8 rounded-md flex justify-center"
            title="添加 Webhook"
          >
            <span class="text-[var(--color-text-primary)]">添加</span>
          </BaseButton>
        </div>
      </div>

      <!-- Webhook 列表 -->
      <div v-else>
        <div v-if="Webhooks.length === 0" class="flex flex-col items-center justify-center mt-2">
          <span class="text-[var(--color-text-muted)]">暂无 Webhook...</span>
        </div>

        <div v-else class="mt-2 overflow-x-auto border border-[var(--color-border-subtle)] rounded-lg">
          <table class="min-w-full divide-y divide-[var(--color-border-subtle)]">
            <thead>
              <tr class="bg-[var(--color-bg-surface)] opacity-70">
                <th
                  class="px-3 py-2 text-left text-sm font-semibold text-[var(--color-text-primary)]"
                >
                  名称
                </th>
                <th
                  class="px-3 py-2 text-left text-sm font-semibold text-[var(--color-text-primary)]"
                >
                  URL
                </th>
                <th
                  class="px-3 py-2 text-right text-sm font-semibold text-[var(--color-text-primary)]"
                >
                  操作
                </th>
              </tr>
            </thead>
            <tbody class="divide-y divide-[var(--color-border-subtle)] text-nowrap">
              <tr v-for="webhook in Webhooks" :key="webhook.id">
                <td class="px-3 py-2 text-sm text-[var(--color-text-primary)]">
                  <span :title="webhook.name" class="truncate block max-w-xs">{{
                    webhook.name
                  }}</span>
                </td>
                <td
                  class="px-3 py-2 text-sm text-[var(--color-text-primary)] font-mono truncate max-w-xs"
                  :title="webhook.url"
                >
                  {{ webhook.url }}
                </td>
                <td class="px-3 py-2 text-right">
                  <button
                    class="p-1 hover:bg-[var(--color-bg-surface)] rounded"
                    @click="handleDeleteWebhook(webhook.id)"
                    title="删除 Webhook"
                  >
                    <Trashbin class="w-5 h-5 text-[var(--color-danger)]" />
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </PanelCard>
</template>

<script setup lang="ts">
import PanelCard from '@/layout/PanelCard.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseEditCapsule from '@/components/common/BaseEditCapsule.vue'
import Trashbin from '@/components/icons/trashbin.vue'
import { ref, onMounted } from 'vue'
import { useSettingStore } from '@/stores'
import { storeToRefs } from 'pinia'
import { fetchDeleteWebhook, fetchCreateWebhook } from '@/service/api'
import { useBaseDialog } from '@/composables/useBaseDialog'
import { theToast } from '@/utils/toast'

const webhookEdit = ref<boolean>(false)

const settingStore = useSettingStore()
const { Webhooks } = storeToRefs(settingStore)
const { openConfirm } = useBaseDialog()

const webhookToAdd = ref<App.Api.Setting.WebhookDto>({
  name: '',
  url: '',
  is_active: true,
})
const isSubmitting = ref<boolean>(false)

const handleAddWebhook = () => {
  if (isSubmitting.value) return
  isSubmitting.value = true

  if (!webhookToAdd.value?.name || !webhookToAdd.value?.url) {
    theToast.error('请填写完整的 Webhook 信息')
    isSubmitting.value = false
    return
  }

  fetchCreateWebhook(webhookToAdd.value)
    .then((res) => {
      if (res.code === 1) {
        theToast.success('添加 Webhook 成功')
        webhookToAdd.value = { name: '', url: '', is_active: true }
        settingStore.getAllWebhooks()
      }

      isSubmitting.value = false

      handleCancelAddWebhook()
    })
    .finally(() => {
      isSubmitting.value = false
    })
}

const handleCancelAddWebhook = () => {
  webhookEdit.value = false
  webhookToAdd.value = { name: '', url: '', is_active: true }
}

const handleDeleteWebhook = (id: string) => {
  openConfirm({
    title: '确认删除此 Webhook 吗？',
    description: '删除后将无法恢复，请谨慎操作！',
    onConfirm: () => {
      fetchDeleteWebhook(id).then((res) => {
        if (res.code === 1) {
          theToast.success('删除 Webhook 成功')
          settingStore.getAllWebhooks()
        }
      })
    },
  })
}

onMounted(async () => {
  await settingStore.getAllWebhooks()
})
</script>

<style scoped></style>
