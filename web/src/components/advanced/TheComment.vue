<template>
  <div id="comments" class="w-full max-w-sm h-auto px-0 py-4 my-4 mx-auto">
    <div
      v-if="formMeta && !formMeta.enable_comment"
      class="rounded-lg border border-[var(--color-border-subtle)] p-3 text-sm text-[var(--color-text-muted)]"
    >
      评论功能未开启。
    </div>

    <template v-else>
    <div class="mb-4 comment-list-board">
      <div class="mb-2 flex items-center justify-between">
        <h3 class="font-semibold text-[var(--color-text-primary)]">评论</h3>
        <span class="text-xs text-[var(--color-text-muted)]">{{ comments.length }} 条</span>
      </div>

      <div v-if="loading" class="text-sm text-[var(--color-text-muted)]">评论加载中...</div>
      <div
        v-else-if="comments.length === 0"
        class="text-sm text-[var(--color-text-muted)]"
      >
        暂无评论，来做第一个留言的人吧。
      </div>
      <div v-else class="space-y-4 pt-1">
        <article
          v-for="(item, index) in comments"
          :key="item.id"
          class="comment-sticky relative rounded-[4px] border p-3"
          :style="getStickyCardStyle(index)"
        >
          <div class="mb-2 flex items-center gap-2">
            <img :src="resolveCommentAvatar(item, index)" alt="avatar" class="h-8 w-8 rounded-full object-cover" />
            <div class="min-w-0">
              <div class="flex items-center gap-1 min-w-0">
                <a
                  v-if="item.website"
                  :href="item.website"
                  target="_blank"
                  rel="noreferrer"
                  class="comment-author-link truncate text-sm font-medium text-[var(--color-text-primary)]"
                >
                  {{ item.nickname }}
                </a>
                <div v-else class="truncate text-sm font-medium text-[var(--color-text-primary)]">
                  {{ item.nickname }}
                </div>
                <Verified
                  v-if="item.source === 'system'"
                  class="verified-badge-icon h-3.5! w-3.5! shrink-0 text-sky-500"
                />
              </div>
              <div class="text-xs text-[var(--color-text-muted)]">
                {{ formatDate(item.created_at) }}
              </div>
            </div>
          </div>
          <TheMdPreview class="comment-md-content" :content="item.content" />
        </article>
      </div>
    </div>

    <form class="comment-form-panel rounded-lg border border-[var(--color-border-subtle)] p-3" @submit.prevent="submitComment">
      <div class="mb-2 flex items-center justify-between">
        <h3 class="font-semibold text-[var(--color-text-primary)]">发表评论</h3>
        <span class="comment-ready-indicator">
          <i
            class="comment-ready-dot"
            :class="profileReady ? 'is-ready' : 'is-not-ready'"
            aria-hidden="true"
          ></i>
        </span>
      </div>

      <article
        class="comment-sticky comment-preview relative mb-2 rounded-[4px] border p-3"
        :style="getStickyCardStyle(0)"
      >
        <div class="mb-2 flex items-center gap-2">
          <img :src="previewAvatar" alt="avatar" class="h-8 w-8 rounded-full object-cover" />
          <div class="min-w-0">
            <div class="truncate text-sm font-medium text-[var(--color-text-primary)]">
              {{ previewNickname }}
            </div>
            <div class="text-xs text-[var(--color-text-muted)]">实时预览</div>
          </div>
        </div>
        <TheMdPreview
          class="comment-md-content"
          :content="form.content || '在这里输入评论内容，将实时预览贴纸效果。'"
        />
      </article>

      <div v-if="!isPrivilegedUser" class="space-y-2">
        <input
          v-model.trim="form.nickname"
          type="text"
          class="w-full rounded-md border border-[var(--color-border-subtle)] bg-transparent px-3 py-2 text-sm"
          placeholder="昵称（必填）"
        />
        <input
          v-model.trim="form.email"
          type="email"
          class="w-full rounded-md border border-[var(--color-border-subtle)] bg-transparent px-3 py-2 text-sm"
          placeholder="邮箱（必填）"
        />
        <input
          v-model.trim="form.website"
          type="url"
          class="w-full rounded-md border border-[var(--color-border-subtle)] bg-transparent px-3 py-2 text-sm"
          placeholder="网址（可选）"
        />
      </div>

      <textarea
        v-model.trim="form.content"
        class="mt-2 min-h-24 w-full rounded-md border border-[var(--color-border-subtle)] bg-transparent px-3 py-2 text-sm"
        placeholder="写下你的评论..."
        maxlength="200"
      />
      <div class="mt-1 text-right text-xs" :class="contentTooLong ? 'text-red-500' : 'text-[var(--color-text-muted)]'">
        {{ contentLength }}/200
      </div>

      <input
        v-model="form.hp_field"
        type="text"
        tabindex="-1"
        autocomplete="off"
        class="hidden"
        aria-hidden="true"
      />

      <input
        v-if="formMeta?.captcha_enabled"
        v-model.trim="form.captcha_token"
        type="text"
        class="mt-2 w-full rounded-md border border-[var(--color-border-subtle)] bg-transparent px-3 py-2 text-sm"
        placeholder="请输入验证码 token（已开启验证码）"
      />

      <div class="mt-3 flex items-center justify-end">
        <button
          type="submit"
          class="rounded-md bg-[var(--color-text-primary)] px-4 py-2 text-sm text-[var(--color-bg-canvas)]"
          :disabled="submitting || !canSubmit"
        >
          {{ submitting ? '提交中...' : '提交评论' }}
        </button>
      </div>
    </form>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import { fetchCreateComment, fetchGetCommentFormMeta, fetchGetComments } from '@/service/api'
import { useUserStore } from '@/stores'
import { theToast } from '@/utils/toast'
import { formatDate } from '@/utils/other'
import TheMdPreview from './TheMdPreview.vue'
import Verified from '../icons/verified.vue'

const route = useRoute()
const userStore = useUserStore()
const loading = ref(false)
const submitting = ref(false)
const comments = ref<App.Api.Comment.CommentItem[]>([])
const formMeta = ref<App.Api.Comment.FormMeta | null>(null)

const form = reactive<App.Api.Comment.CreateCommentDto>({
  echo_id: '',
  nickname: '',
  email: '',
  website: '',
  content: '',
  hp_field: '',
  form_token: '',
  captcha_token: '',
})

const isPrivilegedUser = computed(() => {
  const user = userStore.user
  return Boolean(user && (user.is_admin || user.is_owner))
})

const canSubmit = computed(() => {
  if (!form.echo_id || !form.content) return false
  if (contentTooLong.value) return false
  if (!form.form_token) return false
  if (!isPrivilegedUser.value && (!form.nickname || !form.email)) return false
  if (formMeta.value?.captcha_enabled && !form.captcha_token) return false
  return true
})

const contentLength = computed(() => form.content.length)
const contentTooLong = computed(() => contentLength.value > 200)

const profileReady = computed(() => {
  if (isPrivilegedUser.value) return true
  return Boolean(form.nickname && form.email)
})

const previewNickname = computed(() => {
  if (isPrivilegedUser.value) return userStore.user?.username || '你'
  return form.nickname || '你'
})

const buildDiceBearURL = (seed: string) => {
  const trimmed = seed.trim() || 'guest'
  return `https://api.dicebear.com/9.x/fun-emoji/svg?seed=${encodeURIComponent(trimmed)}`
}

const previewAvatar = computed(() => {
  if (isPrivilegedUser.value && userStore.user?.avatar) {
    return userStore.user.avatar
  }
  return buildDiceBearURL(`${form.nickname || 'guest'}-${form.email || 'preview'}`)
})

const resolveCommentAvatar = (item: App.Api.Comment.CommentItem, index: number) => {
  if (item.source === 'system' && item.avatar_url) return item.avatar_url
  return buildDiceBearURL(`${item.id}-${item.nickname}-${index}`)
}

const getStickyCardStyle = (index: number) => {
  const rotateOptions = ['-0.25deg', '0deg', '0.2deg', '-0.15deg']
  const shiftOptions = ['-1px', '0px', '1px', '0px']
  return {
    '--sticky-rotate': rotateOptions[index % rotateOptions.length],
    '--sticky-shift': shiftOptions[index % shiftOptions.length],
  } as Record<string, string>
}

const loadData = async () => {
  const echoId = String(route.params.echoId || '')
  if (!echoId) return
  form.echo_id = echoId
  loading.value = true
  try {
    const [metaRes, commentsRes] = await Promise.all([
      fetchGetCommentFormMeta(),
      fetchGetComments(echoId),
    ])
    if (metaRes.code === 1) {
      formMeta.value = metaRes.data
      form.form_token = metaRes.data.form_token
    }
    if (commentsRes.code === 1) {
      comments.value = commentsRes.data || []
    }
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  form.content = ''
  form.hp_field = ''
  form.captcha_token = ''
}

const submitComment = async () => {
  if (!canSubmit.value || submitting.value) return
  submitting.value = true
  try {
    const res = await fetchCreateComment(form)
    if (res.code === 1) {
      theToast.success('评论提交成功')
      resetForm()
      await loadData()
    }
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.comment-list-board {
  position: relative;
  padding: 0.75rem;
  border-radius: 10px;
}

:global(body) {
  background-color: color-mix(in srgb, var(--color-bg-canvas) 88%, #f3f2ee 12%);
  background-image:
    linear-gradient(to right, rgba(120, 120, 120, 0.08) 1px, transparent 1px),
    linear-gradient(to bottom, rgba(120, 120, 120, 0.08) 1px, transparent 1px);
  background-size: 32px 32px;
}

.comment-sticky {
  border-color: color-mix(in srgb, var(--color-border-subtle) 78%, #d4c28f 22%);
  background: linear-gradient(180deg, #fffdf3 0%, #fffbed 100%);
  box-shadow:
    0 1px 0 rgba(20, 20, 20, 0.05),
    0 8px 14px rgba(20, 20, 20, 0.06);
  transform: translateX(var(--sticky-shift, 0px)) rotate(var(--sticky-rotate, 0deg));
  transform-origin: 42% 8%;
  border-radius: 4px;
}

.comment-sticky::after {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.26) 0%, rgba(255, 255, 255, 0) 44%);
  pointer-events: none;
}

.comment-sticky::before {
  content: '';
  position: absolute;
  inset: 0;
  background-image:
    radial-gradient(rgba(116, 97, 56, 0.15) 0.65px, transparent 0.8px),
    radial-gradient(rgba(78, 63, 30, 0.08) 0.6px, transparent 0.9px),
    radial-gradient(rgba(255, 255, 255, 0.18) 0.55px, transparent 0.7px),
    linear-gradient(115deg, rgba(132, 107, 62, 0.05) 0%, rgba(132, 107, 62, 0) 45%);
  background-position:
    0 0,
    1px 2px,
    3px 1px,
    0 0;
  background-size:
    5px 5px,
    7px 6px,
    4px 4px,
    100% 100%;
  opacity: 0.5;
  pointer-events: none;
}

.comment-author-link {
  display: inline-block;
  max-width: 100%;
  transition: color 0.2s ease;
}

.comment-author-link:hover {
  color: #0ea5e9;
}

.verified-badge-icon {
  transform: translateY(0.5px);
}

.comment-ready-indicator {
  display: inline-flex;
  align-items: center;
}

.comment-ready-dot {
  width: 0.55rem;
  height: 0.55rem;
  border-radius: 9999px;
  box-shadow: 0 0 0 1px rgba(255, 255, 255, 0.72);
}

.comment-ready-dot.is-ready {
  background: #10b981;
}

.comment-ready-dot.is-not-ready {
  background: #ef4444;
}

.comment-form-panel {
  background: color-mix(in srgb, var(--color-bg-canvas) 92%, #fff 8%);
  box-shadow:
    0 1px 0 rgba(20, 20, 20, 0.04),
    0 10px 18px rgba(20, 20, 20, 0.08);
  border-color: color-mix(in srgb, var(--color-border-subtle) 78%, #cabd95 22%);
}

.comment-md-content {
  color: var(--color-text-primary);
  font-size: 0.9rem;
  line-height: 1.65;
}

.comment-md-content :deep(p) {
  margin: 0.15rem 0;
  white-space: pre-wrap;
  word-break: break-word;
}

.comment-preview {
  opacity: 0.98;
}

@media (max-width: 640px) {
  .comment-sticky {
    transform: translateX(calc(var(--sticky-shift, 0px) * 0.35)) rotate(calc(var(--sticky-rotate, 0deg) * 0.35));
    box-shadow:
      0 1px 0 rgba(20, 20, 20, 0.06),
      0 8px 14px rgba(20, 20, 20, 0.08);
  }
}
</style>
