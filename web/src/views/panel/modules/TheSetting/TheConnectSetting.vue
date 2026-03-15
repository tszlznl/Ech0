<template>
  <PanelCard>
    <!-- Ech0 Connect设置 -->
    <div class="w-full">
      <div class="flex flex-row items-center justify-between mb-3">
        <h1 class="text-[var(--color-text-primary)] font-bold text-lg">Ech0 Connect</h1>
        <div class="flex flex-row items-center justify-end">
          <BaseEditCapsule
            :editing="connectsEdit"
            :apply-title="t('commonUi.done')"
            :cancel-title="t('commonUi.cancel')"
            :edit-title="t('commonUi.edit')"
            @apply="connectsEdit = false"
            @toggle="connectsEdit = !connectsEdit"
          />
        </div>
      </div>

      <!-- 添加 Connect -->
      <div v-if="connectsEdit" class="text-[var(--color-text-secondary)] mb-2">
        <div class="flex items-center gap-2">
          <BaseInput
            v-model="connectUrl"
            type="text"
            :placeholder="t('connectSetting.urlPlaceholder')"
            class="flex-1 h-8"
          />
          <BaseButton
            :icon="Publish"
            @click="handleAddConnect"
            class="w-8 h-8 rounded-md"
            :title="t('connectSetting.connect')"
          />
        </div>
      </div>

      <!-- Connect 列表 -->
      <div v-else>
        <div v-if="connects.length === 0" class="flex flex-col items-center justify-center mt-2">
          <span class="text-[var(--color-text-muted)]">{{ t('connectSetting.empty') }}</span>
        </div>

        <div
          v-else
          class="mt-2 x-scrollbar overflow-x-auto border border-[var(--color-border-subtle)] rounded-lg"
        >
          <table class="min-w-full divide-y divide-[var(--color-border-subtle)]">
            <thead>
              <tr class="bg-[var(--color-bg-surface)] opacity-70">
                <th
                  class="px-3 py-2 text-left text-sm font-semibold text-[var(--color-text-primary)]"
                >
                  #
                </th>
                <th
                  class="px-3 py-2 text-left text-sm font-semibold text-[var(--color-text-primary)]"
                >
                  {{ t('connectSetting.connectUrl') }}
                </th>
                <th
                  class="px-3 min-w-18 py-2 text-right text-sm font-semibold text-[var(--color-text-primary)]"
                >
                  {{ t('commonUi.actions') }}
                </th>
              </tr>
            </thead>
            <tbody class="divide-y divide-[var(--color-border-subtle)] text-nowrap">
              <tr v-for="(connect, index) in connects" :key="connect.id">
                <td class="px-3 py-2 text-sm text-[var(--color-text-primary)]">{{ index + 1 }}</td>
                <td
                  class="px-3 py-2 text-sm text-[var(--color-text-primary)] font-mono truncate max-w-xs"
                  :title="connect.connect_url"
                >
                  {{ connect.connect_url }}
                </td>
                <td class="px-3 py-2 text-right">
                  <button
                    class="p-1 hover:bg-[var(--color-bg-surface)] rounded"
                    @click="handleDisconnect(connect.id)"
                    :title="t('connectSetting.disconnect')"
                  >
                    <Disconnect class="w-5 h-5 text-[var(--color-danger)]" />
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
import Disconnect from '@/components/icons/disconnect.vue'
import Publish from '@/components/icons/publish.vue'
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { fetchAddConnect, fetchDeleteConnect } from '@/service/api'
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

const handleAddConnect = async () => {
  if (connectUrl.value.length === 0) {
    theToast.error(String(t('connectSetting.enterAddress')))
    return
  }
  await fetchAddConnect(connectUrl.value).then((res) => {
    if (res.code === 1) {
      theToast.success(res.msg)
      connectUrl.value = ''
      getConnect()
    }
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
          getConnect()
        }
      })
    },
  })
}

onMounted(() => {
  getConnect()
})
</script>

<style scoped></style>
