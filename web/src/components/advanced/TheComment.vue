<template>
  <div
    v-show="shouldShowComment"
    ref="rootRef"
    id="comments"
    class="max-w-sm h-auto p-4 my-4 mx-auto"
  >
    <div ref="commentRef"></div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onBeforeUnmount, nextTick } from 'vue'
import { useSettingStore } from '@/stores'
import { storeToRefs } from 'pinia'
import { createCommentAdapter } from './comment-providers/registry'
import type { CommentProviderAdapter } from './comment-providers/types'

const { CommentSetting, loading } = storeToRefs(useSettingStore())
const rootRef = ref<HTMLElement | null>(null)
const commentRef = ref<HTMLElement | null>(null)
const canMount = ref(false)
const shouldShowComment = computed(() => CommentSetting.value?.enable_comment)

let observer: IntersectionObserver | null = null
let adapter: CommentProviderAdapter | null = null
let rerenderTimer: ReturnType<typeof setTimeout> | null = null
let renderToken = 0

const setupObserver = async () => {
  await nextTick()
  if (!rootRef.value) return
  observer?.disconnect()
  observer = new IntersectionObserver(
    (entries) => {
      const first = entries[0]
      if (first?.isIntersecting) {
        canMount.value = true
      }
    },
    {
      rootMargin: '240px 0px 240px 0px',
      threshold: 0.01,
    },
  )
  observer.observe(rootRef.value)
}

const unmountAdapter = async () => {
  await adapter?.unmount?.()
  adapter = null
  if (commentRef.value) {
    commentRef.value.innerHTML = ''
  }
}

const mountAdapter = async () => {
  if (loading.value || !shouldShowComment.value || !canMount.value || !commentRef.value) {
    return
  }
  const currentToken = ++renderToken
  const provider = CommentSetting.value.provider
  const nextAdapter = await createCommentAdapter(provider)
  if (!nextAdapter || currentToken !== renderToken) return

  await unmountAdapter()
  adapter = nextAdapter
  await adapter.mount(commentRef.value, CommentSetting.value)
}

const scheduleRender = () => {
  if (rerenderTimer) {
    clearTimeout(rerenderTimer)
  }
  rerenderTimer = setTimeout(() => {
    mountAdapter()
  }, 120)
}

watch(
  () => shouldShowComment.value,
  async (show) => {
    if (!show) {
      await unmountAdapter()
      return
    }
    setupObserver()
    scheduleRender()
  },
  { immediate: true },
)

watch(
  () => canMount.value,
  (ready) => {
    if (ready) scheduleRender()
  },
)

watch(
  () => [
    loading.value,
    CommentSetting.value.provider,
    JSON.stringify(CommentSetting.value.providers),
  ],
  () => {
    scheduleRender()
  },
)

onBeforeUnmount(async () => {
  if (rerenderTimer) clearTimeout(rerenderTimer)
  observer?.disconnect()
  await unmountAdapter()
})
</script>
