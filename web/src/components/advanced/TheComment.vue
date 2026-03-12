<template>
  <div id="comments" class="w-full max-w-sm h-auto px-0 py-4 my-4 mx-auto">
    <div
      v-if="formMeta && !formMeta.enable_comment"
      class="rounded-lg border border-[var(--color-border-subtle)] p-3 text-sm text-[var(--color-text-muted)]"
    >
      评论功能未开启。
    </div>

    <template v-else>
    <div class="mb-4">
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
          class="comment-sticky relative rounded-md border p-3"
          :style="getStickyCardStyle(index)"
        >
          <span class="sticky-pin" aria-hidden="true"></span>
          <div class="mb-2 flex items-center gap-2">
            <img :src="resolveCommentAvatar(item, index)" alt="avatar" class="h-8 w-8 rounded-full object-cover" />
            <div class="min-w-0">
              <div class="truncate text-sm font-medium text-[var(--color-text-primary)]">
                {{ item.nickname }}
              </div>
              <div class="text-xs text-[var(--color-text-muted)]">
                {{ formatDate(item.created_at) }}
              </div>
            </div>
            <a
              v-if="item.website"
              :href="item.website"
              target="_blank"
              rel="noreferrer"
              class="ml-auto text-xs text-sky-500 hover:underline"
            >
              主页
            </a>
          </div>
          <TheMdPreview class="comment-md-content" :content="item.content" />
        </article>
      </div>
    </div>

    <form class="rounded-lg border border-[var(--color-border-subtle)] p-3" @submit.prevent="submitComment">
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
        class="comment-sticky comment-preview relative mb-2 rounded-md border p-3"
        :style="getStickyCardStyle(0)"
      >
        <span class="sticky-pin" aria-hidden="true"></span>
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
  const rotateOptions = ['-0.8deg', '0.5deg', '-0.3deg', '0.9deg']
  const shiftOptions = ['-4px', '2px', '-1px', '3px']
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
.comment-sticky {
  border-color: color-mix(in srgb, var(--color-border-subtle) 70%, #c7b88a 30%);
  background: linear-gradient(180deg, #fffdf6 0%, #fff9ea 100%);
  box-shadow:
    0 1px 0 rgba(20, 20, 20, 0.08),
    0 10px 18px rgba(20, 20, 20, 0.08);
  transform: translateX(var(--sticky-shift, 0px)) rotate(var(--sticky-rotate, 0deg));
  transform-origin: 50% 8%;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.comment-sticky::after {
  content: '';
  position: absolute;
  top: 0;
  right: 0;
  width: 22px;
  height: 22px;
  background: linear-gradient(135deg, #f4e5b4 0%, #f1dd9d 60%, #e4cb7e 100%);
  clip-path: polygon(0 0, 100% 0, 100% 100%);
  opacity: 0.72;
}

.comment-sticky:hover {
  transform: translateX(var(--sticky-shift, 0px)) rotate(var(--sticky-rotate, 0deg)) translateY(-2px);
  box-shadow:
    0 2px 0 rgba(20, 20, 20, 0.08),
    0 14px 26px rgba(20, 20, 20, 0.12);
}

.sticky-pin {
  position: absolute;
  top: 8px;
  left: 50%;
  width: 10px;
  height: 10px;
  border-radius: 9999px;
  background: radial-gradient(circle at 35% 35%, #fff7d5 0%, #d5bd7f 60%, #9f8448 100%);
  box-shadow:
    0 1px 1px rgba(0, 0, 0, 0.24),
    0 0 0 2px rgba(255, 255, 255, 0.78);
  transform: translateX(-50%);
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
  opacity: 0.96;
}

@media (max-width: 640px) {
  .comment-sticky {
    transform: translateX(calc(var(--sticky-shift, 0px) * 0.35)) rotate(calc(var(--sticky-rotate, 0deg) * 0.35));
    box-shadow:
      0 1px 0 rgba(20, 20, 20, 0.06),
      0 8px 14px rgba(20, 20, 20, 0.08);
  }

  .comment-sticky:hover {
    transform: translateX(calc(var(--sticky-shift, 0px) * 0.35)) rotate(calc(var(--sticky-rotate, 0deg) * 0.35));
  }
}
</style>
