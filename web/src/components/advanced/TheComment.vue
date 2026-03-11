<template>
  <div
    v-show="shouldShowComment"
    ref="rootRef"
    id="comments"
    class="w-full max-w-sm h-auto px-0 py-4 my-4 mx-auto"
  >
    <div ref="commentRef"></div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onBeforeUnmount, nextTick } from 'vue'
import { useSettingStore } from '@/stores'
import { storeToRefs } from 'pinia'
import { CommentProvider } from '@/enums/enums'
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
const TWIKOO_FORCE_STYLE_ID = 'twikoo-force-three-row-style'

const ensureTwikooThreeRowStyle = () => {
  if (document.getElementById(TWIKOO_FORCE_STYLE_ID)) return
  const style = document.createElement('style')
  style.id = TWIKOO_FORCE_STYLE_ID
  style.textContent = `
#comments #twikoo-comment-container .tk-submit .tk-meta-input {
  display: flex !important;
  flex-direction: column !important;
  align-items: stretch !important;
  width: 100% !important;
  max-width: 100% !important;
}
#comments #twikoo-comment-container .tk-submit .tk-meta-input .el-input {
  width: 100% !important;
  max-width: 100% !important;
  flex: 0 0 100% !important;
  margin-left: 0 !important;
}
#comments #twikoo-comment-container .tk-submit .tk-meta-input .el-input + .el-input {
  margin-top: 10px !important;
}
`
  document.head.appendChild(style)
}

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

  try {
    await unmountAdapter()
    adapter = nextAdapter
    await adapter.mount(commentRef.value, CommentSetting.value)
    if (provider === CommentProvider.TWIKOO) {
      ensureTwikooThreeRowStyle()
    }
  } catch (error) {
    console.error('[comment] mount adapter failed:', error)
  }
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

<style scoped>
#comments :deep(#twikoo-comment-container) {
  max-width: 100%;
  overflow-x: hidden;
}

#comments :deep(#twikoo-comment-container .tk-submit),
#comments :deep(#twikoo-comment-container .tk-comments) {
  max-width: 100%;
}

#comments :deep(#twikoo-comment-container .tk-row) {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  width: 100%;
  max-width: 100%;
  margin-left: 0;
  margin-right: 0;
}

#comments :deep(#twikoo-comment-container .tk-avatar) {
  flex: 0 0 36px;
  width: 36px;
  min-width: 36px;
}

#comments :deep(#twikoo-comment-container .tk-row .tk-col) {
  flex: 1 1 auto;
  min-width: 0;
  width: auto !important;
}

#comments :deep(#twikoo-comment-container .tk-meta-input) {
  display: flex !important;
  flex-direction: column !important;
  align-items: stretch !important;
  gap: 10px;
  width: 100%;
  max-width: 100%;
}

#comments :deep(#twikoo-comment-container .tk-submit .tk-meta-input .el-input) {
  width: 100% !important;
  flex: 0 0 100% !important;
  margin-left: 0 !important;
}

#comments :deep(#twikoo-comment-container .tk-submit .tk-meta-input .el-input + .el-input) {
  margin-left: 0 !important;
  margin-top: 10px !important;
}

#comments :deep(#twikoo-comment-container .tk-meta-input > *) {
  width: 100% !important;
  max-width: 100% !important;
  flex: 0 0 auto !important;
}

#comments :deep(#twikoo-comment-container .tk-meta-input > * + *) {
  margin-left: 0;
  margin-top: 0;
}

#comments :deep(#twikoo-comment-container .tk-meta-input .tk-input) {
  width: 100% !important;
  max-width: 100% !important;
}

#comments :deep(#twikoo-comment-container .tk-meta-input input) {
  width: 100% !important;
  max-width: 100%;
  box-sizing: border-box;
  border: 1px solid rgba(100, 116, 139, 0.35);
  border-radius: 8px;
  background: var(--el-fill-color-blank, #fff);
}

#comments :deep(#twikoo-comment-container .tk-meta-input input::placeholder) {
  color: rgba(100, 116, 139, 0.85);
}

#comments :deep(#twikoo-comment-container .tk-comments .tk-comment) {
  margin-top: 12px;
  padding: 12px 14px;
  border: 1px solid rgba(100, 116, 139, 0.2);
  border-radius: 12px;
  background: rgba(148, 163, 184, 0.08);
}

#comments :deep(#twikoo-comment-container .tk-comment .tk-meta) {
  margin-bottom: 8px;
}

#comments :deep(#twikoo-comment-container .tk-comment .tk-nick) {
  font-weight: 600;
}

#comments :deep(#twikoo-comment-container .tk-comment .tk-time) {
  font-size: 12px;
  opacity: 0.75;
}

#comments :deep(#twikoo-comment-container .tk-comment .tk-content) {
  line-height: 1.75;
}
</style>
