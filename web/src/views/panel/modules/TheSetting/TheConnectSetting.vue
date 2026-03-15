<template>
  <PanelCard>
    <div class="w-full">
      <div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
        <div>
          <h1 class="text-[var(--color-text-primary)] font-bold text-lg">Ech0 Connect</h1>
          <p class="mt-1 text-sm text-[var(--color-text-muted)]">
            {{ t('connectSetting.description') }}
          </p>
        </div>
        <BaseButton
          v-if="!connectsEdit"
          class="top-action-btn top-action-btn-primary shrink-0 whitespace-nowrap px-2.5 py-1 text-xs"
          @click="connectsEdit = true"
        >
          {{ t('connectSetting.addConnect') }}
        </BaseButton>
      </div>

      <div
        v-if="connectsEdit"
        class="mt-4 rounded-lg border border-[var(--color-border-subtle)] bg-[var(--color-bg-surface)]/40 p-4 text-[var(--color-text-secondary)]"
      >
        <div class="mb-1 text-sm text-[var(--color-text-primary)]">
          {{ t('connectSetting.connectUrl') }}
        </div>
        <div class="flex items-center gap-2">
          <BaseInput
            v-model="connectUrl"
            type="text"
            :placeholder="t('connectSetting.urlPlaceholder')"
            class="h-9 flex-1"
          />
        </div>
        <p v-if="connectUrlError" class="mt-1 text-xs text-[var(--color-danger)]">
          {{ connectUrlError }}
        </p>
        <p class="mt-1 text-xs text-[var(--color-text-muted)]">
          {{ t('connectSetting.connectHint') }}
        </p>

        <div class="mt-4 flex flex-col-reverse items-center justify-center gap-2 sm:flex-row">
          <BaseButton
            class="h-9 rounded-md px-4"
            :disabled="isSubmitting"
            @click="handleCancelConnect"
          >
            {{ t('commonUi.cancel') }}
          </BaseButton>
          <BaseButton class="h-9 rounded-md px-4" :loading="isSubmitting" @click="handleAddConnect">
            {{ t('connectSetting.connect') }}
          </BaseButton>
        </div>
      </div>

      <div v-if="connects.length === 0" class="flex flex-col items-center justify-center mt-4">
        <span class="text-[var(--color-text-muted)]">{{ t('connectSetting.empty') }}</span>
      </div>

      <div
        v-else
        class="mt-4 x-scrollbar overflow-x-auto border border-[var(--color-border-subtle)] rounded-lg"
      >
        <table class="w-full min-w-[520px] table-fixed text-sm">
          <thead>
            <tr class="bg-[var(--color-bg-muted)]/70 text-left text-[var(--color-text-muted)]">
              <th class="w-[56px] px-2 py-2 whitespace-nowrap">#</th>
              <th class="px-2 py-2 whitespace-nowrap">
                {{ t('connectSetting.connectUrl') }}
              </th>
              <th class="w-[110px] px-2 py-2 whitespace-nowrap">
                {{ t('connectSetting.status') }}
              </th>
              <th class="w-[88px] px-2 py-2 text-right whitespace-nowrap">
                {{ t('commonUi.actions') }}
              </th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="(connect, index) in connects"
              :key="connect.id"
              class="border-t border-[var(--color-border-subtle)] text-[var(--color-text-secondary)]"
            >
              <td class="px-2 py-2 text-[var(--color-text-primary)]">{{ index + 1 }}</td>
              <td
                class="px-2 py-2 text-[var(--color-text-primary)] font-mono truncate"
                :title="connect.connect_url"
              >
                {{ connect.connect_url }}
              </td>
              <td class="px-1 py-2">
                <span :class="['status-pill', statusClass(connect.id)]">
                  {{ statusLabel(connect.id) }}
                </span>
              </td>
              <td class="px-2 py-2 text-right">
                <BaseButton
                  class="h-8 w-8 !p-1.5"
                  :icon="Disconnect"
                  @click="handleDisconnect(connect.id)"
                  :title="t('connectSetting.disconnect')"
                />
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </PanelCard>
</template>

<script setup lang="ts">
import PanelCard from '@/layout/PanelCard.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import Disconnect from '@/components/icons/disconnect.vue'
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { fetchAddConnect, fetchDeleteConnect, fetchGetConnect } from '@/service/api'
import { theToast } from '@/utils/toast'

import { useConnectStore } from '@/stores'
import { storeToRefs } from 'pinia'

import { useBaseDialog } from '@/composables/useBaseDialog'
const { openConfirm } = useBaseDialog()

const connectStore = useConnectStore()
const { t } = useI18n()
const { getConnect } = connectStore
const { connects } = storeToRefs(connectStore)
const connectsEdit = ref<boolean>(false)
const connectUrl = ref<string>('')
const connectUrlError = ref<string>('')
const isSubmitting = ref<boolean>(false)
const statusById = ref<Record<string, 'checking' | 'online' | 'offline'>>({})

const isValidConnectUrl = (value: string) => {
  try {
    const parsed = new URL(value)
    return parsed.protocol === 'http:' || parsed.protocol === 'https:'
  } catch {
    return false
  }
}

const handleCancelConnect = () => {
  connectUrl.value = ''
  connectUrlError.value = ''
  connectsEdit.value = false
}

const normalizeUrl = (value: string) => {
  try {
    return new URL(value).origin
  } catch {
    return value.trim().replace(/\/+$/, '')
  }
}

const statusClass = (id: string) => {
  const current = statusById.value[id] || 'checking'
  if (current === 'online') return 'status-success'
  if (current === 'offline') return 'status-failed'
  return 'status-checking'
}

const statusLabel = (id: string) => {
  const current = statusById.value[id] || 'checking'
  if (current === 'online') return t('connectSetting.statusOnline')
  if (current === 'offline') return t('connectSetting.statusOffline')
  return t('connectSetting.statusChecking')
}

const refreshConnectivityStatus = async () => {
  const list = [...connects.value]
  const nextState: Record<string, 'checking' | 'online' | 'offline'> = {}
  for (const connect of list) {
    nextState[connect.id] = 'checking'
  }
  statusById.value = nextState

  await Promise.all(
    list.map(async (connect) => {
      try {
        const res = await fetchGetConnect(connect.connect_url, true)
        const inputOrigin = normalizeUrl(connect.connect_url)
        const serverOrigin = normalizeUrl(res?.data?.server_url || '')
        statusById.value[connect.id] =
          res.code === 1 && !!res.data?.server_url && serverOrigin === inputOrigin
            ? 'online'
            : 'offline'
      } catch {
        statusById.value[connect.id] = 'offline'
      }
    }),
  )
}

const refreshConnectData = async () => {
  await getConnect()
  await refreshConnectivityStatus()
}

const handleAddConnect = async () => {
  if (isSubmitting.value) return
  connectUrlError.value = ''
  const target = connectUrl.value.trim()
  if (!target) {
    connectUrlError.value = String(t('connectSetting.enterAddress'))
    return
  }
  if (!isValidConnectUrl(target)) {
    connectUrlError.value = String(t('connectSetting.invalidUrl'))
    return
  }
  isSubmitting.value = true
  await fetchAddConnect(target)
    .then((res) => {
      if (res.code === 1) {
        theToast.success(res.msg)
        connectUrl.value = ''
        connectUrlError.value = ''
        connectsEdit.value = false
        refreshConnectData()
      }
    })
    .finally(() => {
      isSubmitting.value = false
    })
}

const handleDisconnect = async (connect_id: string) => {
  // 弹出确认框
  openConfirm({
    title: String(t('connectSetting.disconnectConfirmTitle')),
    description: '',
    onConfirm: async () => {
      await fetchDeleteConnect(connect_id).then((res) => {
        if (res.code === 1) {
          theToast.success(res.msg)
          refreshConnectData()
        }
      })
    },
  })
}

onMounted(() => {
  refreshConnectData()
})
</script>

<style scoped>
.top-action-btn {
  border: 1px solid var(--color-border-subtle) !important;
  background: var(--color-bg-surface) !important;
  color: var(--color-text-secondary) !important;
}

.top-action-btn:hover {
  border-color: var(--color-border-strong) !important;
  background: var(--color-bg-muted) !important;
}

.top-action-btn-primary {
  border-color: var(--color-border-strong) !important;
}

.status-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 9999px;
  padding: 2px 8px;
  font-size: 12px;
  line-height: 1.2;
  font-weight: 500;
  white-space: nowrap;
  border: 1px solid transparent;
}

.status-success {
  color: #166534;
  border-color: rgba(34, 197, 94, 0.35);
  background: rgba(34, 197, 94, 0.14);
}

.status-failed {
  color: #991b1b;
  border-color: rgba(239, 68, 68, 0.35);
  background: rgba(239, 68, 68, 0.14);
}

.status-checking {
  color: var(--color-text-muted);
  border-color: var(--color-border-subtle);
  background: color-mix(in srgb, var(--color-bg-muted) 65%, transparent);
}
</style>
