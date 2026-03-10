<template>
  <PanelCard>
    <!-- Ech0 Connect设置 -->
    <div class="w-full">
      <div class="flex flex-row items-center justify-between mb-3">
        <h1 class="text-[var(--color-text-primary)] font-bold text-lg">Ech0 Connect</h1>
        <div class="flex flex-row items-center justify-end gap-2 w-14">
          <button @click="connectsEdit = !connectsEdit" title="编辑">
            <Edit
              v-if="!connectsEdit"
              class="w-5 h-5 text-[var(--color-text-muted)] hover:w-6 hover:h-6"
            />
            <Close v-else class="w-5 h-5 text-[var(--color-text-muted)] hover:w-6 hover:h-6" />
          </button>
        </div>
      </div>

      <!-- 添加 Connect -->
      <div v-if="connectsEdit" class="text-[var(--color-text-secondary)] mb-2">
        <div class="flex items-center gap-2">
          <BaseInput
            v-model="connectUrl"
            type="text"
            placeholder="请输入 Connect 地址（带https/http）"
            class="flex-1 h-8"
          />
          <BaseButton
            :icon="Publish"
            @click="handleAddConnect"
            class="w-8 h-8 rounded-md"
            title="连接"
          />
        </div>
      </div>

      <!-- Connect 列表 -->
      <div v-else>
        <div v-if="connects.length === 0" class="flex flex-col items-center justify-center mt-2">
          <span class="text-[var(--color-text-muted)]">暂无连接...</span>
        </div>

        <div v-else class="mt-2 overflow-x-auto border border-[var(--color-border-subtle)] rounded-lg">
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
                  Connect 地址
                </th>
                <th
                  class="px-3 min-w-18 py-2 text-right text-sm font-semibold text-[var(--color-text-primary)]"
                >
                  操作
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
                    title="断开连接"
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
import Edit from '@/components/icons/edit.vue'
import Disconnect from '@/components/icons/disconnect.vue'
import Close from '@/components/icons/close.vue'
import Publish from '@/components/icons/publish.vue'
import { ref, onMounted } from 'vue'
import { fetchAddConnect, fetchDeleteConnect } from '@/service/api'
import { theToast } from '@/utils/toast'

import { useConnectStore } from '@/stores'
import { storeToRefs } from 'pinia'

import { useBaseDialog } from '@/composables/useBaseDialog'
const { openConfirm } = useBaseDialog()

const connectStore = useConnectStore()
const { getConnect } = connectStore
const { connects } = storeToRefs(connectStore)
const connectsEdit = ref<boolean>(false)
const connectUrl = ref<string>('')

const handleAddConnect = async () => {
  if (connectUrl.value.length === 0) {
    theToast.error('请输入Connect地址')
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
    title: '确定要断开连接吗？',
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
