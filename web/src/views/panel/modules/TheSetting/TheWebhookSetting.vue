<template>
  <PanelCard>
    <div class="w-full">
      <div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
        <div>
          <h1 class="text-lg font-bold text-[var(--color-text-primary)]">
            {{ t('webhookSetting.title') }}
          </h1>
          <p class="text-sm text-[var(--color-text-muted)] mt-1">
            {{ t('webhookSetting.description') }}
          </p>
        </div>
        <div class="flex items-center gap-2">
          <BaseButton
            v-if="!isFormOpen"
            class="h-9 rounded-md px-4"
            @click="openCreateForm"
          >
            {{ t('webhookSetting.createWebhook') }}
          </BaseButton>
        </div>
      </div>

      <div
        v-if="isFormOpen"
        class="mt-4 rounded-lg border border-[var(--color-border-subtle)] bg-[var(--color-bg-surface)]/40 p-4"
      >
        <div class="grid gap-3 md:grid-cols-2">
          <div class="md:col-span-1">
            <div class="mb-1 text-sm text-[var(--color-text-primary)]">
              {{ t('webhookSetting.name') }}
            </div>
            <BaseInput
              v-model="webhookForm.name"
              class="w-full"
              :placeholder="t('webhookSetting.namePlaceholder')"
            />
            <p
              v-if="formErrors.name"
              class="mt-1 text-xs text-[var(--color-danger)]"
            >
              {{ formErrors.name }}
            </p>
          </div>

          <div class="md:col-span-1">
            <div class="mb-1 text-sm text-[var(--color-text-primary)]">
              {{ t('webhookSetting.url') }}
            </div>
            <BaseInput
              v-model="webhookForm.url"
              class="w-full font-mono"
              :placeholder="t('webhookSetting.urlPlaceholder')"
            />
            <p
              v-if="formErrors.url"
              class="mt-1 text-xs text-[var(--color-danger)]"
            >
              {{ formErrors.url }}
            </p>
          </div>

          <div class="md:col-span-2">
            <div class="mb-1 text-sm text-[var(--color-text-primary)]">
              {{ t('webhookSetting.secret') }}
            </div>
            <BaseInput
              v-model="webhookForm.secret"
              class="w-full font-mono"
              :placeholder="t('webhookSetting.secretPlaceholder')"
            />
          </div>

          <div class="md:col-span-2 flex items-center justify-between rounded-md border border-[var(--color-border-subtle)] px-3 py-2">
            <div>
              <p class="text-sm text-[var(--color-text-primary)]">
                {{ t('webhookSetting.enableWebhook') }}
              </p>
              <p class="text-xs text-[var(--color-text-muted)]">
                {{ t('webhookSetting.enableWebhookHint') }}
              </p>
            </div>
            <BaseSwitch
              :model-value="webhookForm.is_active"
              @update:model-value="onFormActiveChange"
            />
          </div>
        </div>

        <div class="mt-4 flex flex-col-reverse items-center justify-center gap-2 sm:flex-row">
          <BaseButton
            class="h-9 rounded-md px-4"
            :disabled="formSubmitting"
            @click="resetForm"
          >
            {{ t('commonUi.cancel') }}
          </BaseButton>
          <BaseButton
            class="h-9 rounded-md px-4"
            :loading="formSubmitting"
            @click="submitForm"
          >
            {{ isEditMode ? t('webhookSetting.saveWebhook') : t('webhookSetting.createWebhook') }}
          </BaseButton>
        </div>
      </div>

      <div class="mt-4">
        <div
          v-if="webhooksLoading"
          class="rounded-lg border border-[var(--color-border-subtle)] px-4 py-6 text-center text-sm text-[var(--color-text-muted)]"
        >
          {{ t('webhookSetting.loading') }}
        </div>

        <div
          v-else-if="webhooksError"
          class="rounded-lg border border-[var(--color-border-subtle)] px-4 py-6 text-center"
        >
          <p class="text-sm text-[var(--color-danger)]">
            {{ t('webhookSetting.loadFailed') }}
          </p>
          <p class="mt-1 text-xs text-[var(--color-text-muted)]">
            {{ webhooksError }}
          </p>
          <BaseButton class="mt-3 h-8 rounded-md px-3" @click="refreshWebhooks">
            {{ t('webhookSetting.retryLoad') }}
          </BaseButton>
        </div>

        <div
          v-else-if="Webhooks.length === 0"
          class="rounded-lg border border-[var(--color-border-subtle)] px-4 py-8 text-center text-sm text-[var(--color-text-muted)]"
        >
          {{ t('webhookSetting.empty') }}
        </div>

        <div v-else class="x-scrollbar overflow-x-auto rounded-lg border border-[var(--color-border-subtle)]">
          <table class="w-full min-w-[600px] table-fixed text-sm">
            <thead>
              <tr class="bg-[var(--color-bg-muted)]/70 text-left text-[var(--color-text-muted)]">
                <th class="w-[100px] px-2 py-2 whitespace-nowrap">{{ t('webhookSetting.name') }}</th>
                <th class="w-[190px] px-2 py-2 whitespace-nowrap">URL</th>
                <th class="w-[88px] px-1 py-2 whitespace-nowrap">{{ t('webhookSetting.lastStatus') }}</th>
                <th class="w-[72px] px-1 py-2 whitespace-nowrap">{{ t('webhookSetting.enableWebhook') }}</th>
                <th class="w-[76px] px-1 py-2 text-right whitespace-nowrap">{{ t('commonUi.actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="webhook in Webhooks"
                :key="webhook.id"
                class="border-t border-[var(--color-border-subtle)] text-[var(--color-text-secondary)]"
              >
                <td class="px-2 py-3 text-[var(--color-text-primary)]">
                  {{ webhook.name }}
                </td>
                <td
                  class="max-w-[260px] truncate px-2 py-3 font-mono text-[var(--color-text-primary)]"
                  :title="webhook.url"
                >
                  {{ webhook.url }}
                </td>
                <td class="px-2 py-3">
                  <span class="status-pill" :class="statusClass(webhook.last_status)">
                    {{ statusLabel(webhook.last_status) }}
                  </span>
                </td>
                <td class="px-2 py-3">
                  <BaseSwitch
                    :model-value="webhook.is_active"
                    :disabled="isRowBusy(webhook.id)"
                    @update:model-value="() => handleToggleWebhook(webhook)"
                  />
                </td>
                <td class="px-2 py-3">
                  <div class="flex items-center justify-end gap-1">
                    <BaseButton
                      class="h-8 w-8 !p-1.5"
                      :icon="EditIcon"
                      :disabled="isRowBusy(webhook.id)"
                      :title="t('commonUi.edit')"
                      @click="openEditForm(webhook)"
                    />
                    <BaseButton
                      class="h-8 w-8 !p-1.5"
                      :icon="Trashbin"
                      :disabled="isRowBusy(webhook.id)"
                      :title="t('webhookSetting.delete')"
                      @click="handleDeleteWebhook(webhook.id)"
                    />
                  </div>
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
import BaseButton from '@/components/common/BaseButton.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseSwitch from '@/components/common/BaseSwitch.vue'
import EditIcon from '@/components/icons/edit.vue'
import Trashbin from '@/components/icons/trashbin.vue'
import PanelCard from '@/layout/PanelCard.vue'
import { useBaseDialog } from '@/composables/useBaseDialog'
import { fetchCreateWebhook, fetchDeleteWebhook, fetchUpdateWebhook } from '@/service/api'
import { useSettingStore } from '@/stores'
import { theToast } from '@/utils/toast'
import { storeToRefs } from 'pinia'
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'

type FormMode = 'create' | 'edit'

const { t } = useI18n()
const { openConfirm } = useBaseDialog()
const settingStore = useSettingStore()
const { Webhooks, webhooksError, webhooksLoading } = storeToRefs(settingStore)

const isFormOpen = ref(false)
const formMode = ref<FormMode>('create')
const editingWebhookId = ref<string | null>(null)
const formSubmitting = ref(false)
const togglingId = ref<string | null>(null)
const deletingId = ref<string | null>(null)

const webhookForm = ref<App.Api.Setting.WebhookDto>({
  name: '',
  url: '',
  secret: '',
  is_active: true,
})
const formErrors = ref<{ name: string; url: string }>({
  name: '',
  url: '',
})

const isEditMode = computed(() => formMode.value === 'edit')

const onFormActiveChange = (value: boolean) => {
  webhookForm.value.is_active = value
}

const resetForm = () => {
  isFormOpen.value = false
  formMode.value = 'create'
  editingWebhookId.value = null
  formErrors.value = { name: '', url: '' }
  webhookForm.value = {
    name: '',
    url: '',
    secret: '',
    is_active: true,
  }
}

const openCreateForm = () => {
  resetForm()
  isFormOpen.value = true
}

const openEditForm = (webhook: App.Api.Setting.Webhook) => {
  isFormOpen.value = true
  formMode.value = 'edit'
  editingWebhookId.value = webhook.id
  formErrors.value = { name: '', url: '' }
  webhookForm.value = {
    name: webhook.name,
    url: webhook.url,
    secret: '',
    is_active: webhook.is_active,
  }
}

const refreshWebhooks = async () => {
  await settingStore.getAllWebhooks()
}

const isValidWebhookUrl = (value: string) => {
  try {
    const parsed = new URL(value)
    return parsed.protocol === 'http:' || parsed.protocol === 'https:'
  } catch {
    return false
  }
}

const validateForm = () => {
  formErrors.value = { name: '', url: '' }
  let valid = true
  if (!webhookForm.value.name.trim()) {
    formErrors.value.name = String(t('webhookSetting.fieldNameRequired'))
    valid = false
  }
  if (!webhookForm.value.url.trim()) {
    formErrors.value.url = String(t('webhookSetting.fieldUrlRequired'))
    valid = false
  } else if (!isValidWebhookUrl(webhookForm.value.url)) {
    formErrors.value.url = String(t('webhookSetting.invalidUrl'))
    valid = false
  }
  return valid
}

const submitForm = async () => {
  if (formSubmitting.value) return
  if (!validateForm()) return
  formSubmitting.value = true
  const payload: App.Api.Setting.WebhookDto = {
    name: webhookForm.value.name.trim(),
    url: webhookForm.value.url.trim(),
    secret: webhookForm.value.secret?.trim() || '',
    is_active: webhookForm.value.is_active,
  }
  try {
    const res =
      isEditMode.value && editingWebhookId.value
        ? await fetchUpdateWebhook(editingWebhookId.value, payload)
        : await fetchCreateWebhook(payload)
    if (res.code === 1) {
      theToast.success(
        String(t(isEditMode.value ? 'webhookSetting.updateSuccess' : 'webhookSetting.addSuccess')),
      )
      await refreshWebhooks()
      resetForm()
      return
    }
    theToast.error(String(res.msg || t('webhookSetting.operateFailed')))
  } finally {
    formSubmitting.value = false
  }
}

const isRowBusy = (webhookId: string) =>
  togglingId.value === webhookId || deletingId.value === webhookId

const handleToggleWebhook = async (webhook: App.Api.Setting.Webhook) => {
  if (isRowBusy(webhook.id)) return
  togglingId.value = webhook.id
  try {
    const res = await fetchUpdateWebhook(webhook.id, {
      name: webhook.name,
      url: webhook.url,
      secret: '',
      is_active: !webhook.is_active,
    })
    if (res.code === 1) {
      theToast.success(String(t('webhookSetting.updateSuccess')))
      await refreshWebhooks()
      return
    }
    theToast.error(String(res.msg || t('webhookSetting.operateFailed')))
  } finally {
    togglingId.value = null
  }
}

const handleDeleteWebhook = (id: string) => {
  if (isRowBusy(id)) return
  openConfirm({
    title: String(t('webhookSetting.deleteConfirmTitle')),
    description: String(t('webhookSetting.deleteConfirmDesc')),
    onConfirm: async () => {
      deletingId.value = id
      try {
        const res = await fetchDeleteWebhook(id)
        if (res.code === 1) {
          theToast.success(String(t('webhookSetting.deleteSuccess')))
          await refreshWebhooks()
          return
        }
        theToast.error(String(res.msg || t('webhookSetting.operateFailed')))
      } finally {
        deletingId.value = null
      }
    },
  })
}

const statusLabel = (status: string) => {
  if (status === 'success') return String(t('webhookSetting.statusSuccess'))
  if (status === 'failed') return String(t('webhookSetting.statusFailed'))
  return String(t('webhookSetting.statusUnknown'))
}

const statusClass = (status: string) => {
  if (status === 'success') return 'status-success'
  if (status === 'failed') return 'status-failed'
  return 'status-unknown'
}

onMounted(async () => {
  await refreshWebhooks()
})
</script>

<style scoped>
.status-pill {
  display: inline-flex;
  align-items: center;
  border-radius: 9999px;
  padding: 0.125rem 0.5rem;
  font-size: 0.75rem;
  line-height: 1rem;
}

.status-success {
  color: rgb(34 197 94);
  background: rgb(34 197 94 / 0.12);
}

.status-failed {
  color: rgb(239 68 68);
  background: rgb(239 68 68 / 0.12);
}

.status-unknown {
  color: rgb(245 158 11);
  background: rgb(245 158 11 / 0.14);
}

</style>
