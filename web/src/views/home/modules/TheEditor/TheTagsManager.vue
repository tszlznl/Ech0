<template>
  <div class="py-4">
    <h2 class="text-[var(--color-text-secondary)] font-bold mb-2">
      {{ t('editor.tagManagerTitle') }}
    </h2>
    <p class="text-xs text-[var(--color-text-muted)] mb-3">{{ t('editor.tagManagerHint') }}</p>
    <div class="flex flex-wrap gap-2">
      <Popover v-for="(tag, index) in tagList" :key="tag.id" class="relative overflow-visible" v-slot="{ close }">
        <PopoverButton
          class="flex items-center gap-1 border rounded-sm border-[var(--color-border-subtle)] border-dashed py-0.5 px-1 mb-1 outline-none transition-colors duration-150 hover:text-[var(--color-text-secondary)]"
          style="white-space: nowrap"
        >
          <div
            class="hover:cursor-pointer text-[var(--color-text-muted)] flex items-center justify-start gap-2"
          >
            <div>#</div>
            {{ tag.name }}
          </div>
        </PopoverButton>

        <transition
          enter-active-class="transition duration-200 ease-out"
          enter-from-class="translate-y-1 opacity-0"
          enter-to-class="translate-y-0 opacity-100"
          leave-active-class="transition duration-150 ease-in"
          leave-from-class="translate-y-0 opacity-100"
          leave-to-class="translate-y-1 opacity-0"
        >
          <PopoverPanel :class="getPopoverPanelClass(index)">
            <div class="overflow-hidden rounded-lg shadow-lg ring-1 ring-black/5">
              <div
                class="relative flex justify-around gap-2 bg-[var(--color-bg-surface)] p-1 text-[var(--color-text-secondary)]"
              >
                <button
                  @click="
                    () => {
                      handleFilterByTag(tag)
                      close()
                    }
                  "
                  v-tooltip="t('editor.filterByTag')"
                  class="flex items-center justify-center rounded-md p-1 transition duration-150 ease-in-out hover:text-[var(--color-text-primary)] focus:outline-none focus-visible:ring focus-visible:ring-[var(--input-focus-color-border-subtle)]"
                >
                  <Filter class="w-5 h-5" />
                </button>
                <div v-if="isLogin" class="w-px bg-[var(--color-bg-muted)]"></div>
                <button
                  v-if="isLogin"
                  @click="
                    () => {
                      handleDeleteTag(tag.id)
                      close()
                    }
                  "
                  v-tooltip="t('editor.deleteTag')"
                  class="flex items-center justify-center rounded-md p-1 transition duration-150 ease-in-out hover:text-[var(--color-danger)] focus:outline-none focus-visible:ring focus-visible:ring-[var(--input-focus-color-border-subtle)]"
                >
                  <Trashbin class="w-5 h-5" />
                </button>
              </div>
            </div>
          </PopoverPanel>
        </transition>
      </Popover>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useEchoStore, useUserStore } from '@/stores'
import { fetchDeleteTagById } from '@/service/api'
import { storeToRefs } from 'pinia'
import { useBaseDialog } from '@/composables/useBaseDialog'
import { Popover, PopoverButton, PopoverPanel } from '@headlessui/vue'
import Trashbin from '@/components/icons/trashbin.vue'
import Filter from '@/components/icons/filter.vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'

const echoStore = useEchoStore()
const userStore = useUserStore()
const { tagList } = storeToRefs(echoStore)
const { isLogin } = storeToRefs(userStore)
const { t } = useI18n()
const router = useRouter()

const { openConfirm } = useBaseDialog()
const getPopoverPanelClass = (index: number) => {
  const total = tagList.value.length
  if (index <= 1) return 'absolute left-0 z-40 mt-1'
  if (index >= total - 2) return 'absolute right-0 z-40 mt-1'
  return 'absolute left-1/2 z-40 mt-1 -translate-x-1/2 transform'
}

// 按标签过滤内容
const handleFilterByTag = (tag: App.Api.Ech0.Tag) => {
  if (!tag) return

  echoStore.filteredTag = tag
  echoStore.isFilteringMode = true
  router.push({ name: 'home' })
}

// 删除标签
const handleDeleteTag = (tagId: string) => {
  openConfirm({
    title: String(t('editor.deleteTagConfirmTitle')),
    description: String(t('editor.deleteTagConfirmDesc')),
    onConfirm: () => {
      fetchDeleteTagById(tagId).then((res) => {
        if (res.code === 1) {
          echoStore.getTags()
        }
      })
    },
  })
}
</script>

<style scoped></style>
