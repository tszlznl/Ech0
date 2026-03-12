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
        <div v-else-if="comments.length === 0" class="text-sm text-[var(--color-text-muted)]">
          暂无评论，来做第一个留言的人吧。
        </div>
        <div v-else class="space-y-4 pt-1">
          <article
            v-for="(item, index) in comments"
            :key="item.id"
            class="comment-sticky relative rounded-[4px] border p-3"
            :style="getStickyCardStyle(index)"
          >
            <span v-if="item.hot" class="comment-hot-badge">Hot</span>
            <div class="mb-2 flex items-center gap-2">
              <img
                :src="resolveCommentAvatar(item, index)"
                alt="avatar"
                class="h-8 w-8 rounded-full object-cover"
              />
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

      <button
        v-if="!commentFormExpanded"
        type="button"
        class="comment-pill-btn"
        @click="commentFormExpanded = true"
      >
        <span class="comment-pill-btn__icon" aria-hidden="true">+</span>
        <span>发表评论</span>
      </button>

      <form
        v-else
        class="comment-form-panel rounded-lg border border-[var(--color-border-subtle)] p-3"
        @submit.prevent="submitComment"
      >
        <div class="mb-2 flex items-center justify-between">
          <div class="flex items-center gap-1.5">
            <h3 class="font-semibold text-[var(--color-text-primary)]">发表评论</h3>
            <span
              class="inline-flex items-center gap-1 rounded-full border border-[var(--color-border-subtle)] px-2 py-[2px] text-[11px] text-[var(--color-text-muted)]"
              title="支持 Markdown 语法"
            >
              <MarkdownIcon class="h-3.5 w-3.5" />
              <span>Markdown</span>
            </span>
          </div>
          <div class="flex items-center gap-2">
            <span class="comment-ready-indicator">
              <i
                class="comment-ready-dot"
                :class="profileReady ? 'is-ready' : 'is-not-ready'"
                aria-hidden="true"
              ></i>
            </span>
            <button
              type="button"
              class="comment-collapse-btn"
              aria-label="收起发表评论"
              @click="commentFormExpanded = false"
            >
              收起
            </button>
          </div>
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
        <div
          class="mt-1 text-right text-xs"
          :class="contentTooLong ? 'text-red-500' : 'text-[var(--color-text-muted)]'"
        >
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

        <div class="comment-submit-row mt-3">
          <div v-if="needCaptcha" class="comment-captcha-wrap">
            <div ref="captchaMountRef" class="comment-captcha-mount"></div>
            <p v-if="captchaError" class="comment-captcha-error text-xs text-red-500">
              {{ captchaError }}
            </p>
          </div>
          <button
            type="submit"
            class="comment-submit-btn rounded-md bg-[var(--color-text-primary)] px-4 py-1 text-sm text-[var(--color-bg-canvas)]"
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
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { fetchCreateComment, fetchGetCommentFormMeta, fetchGetComments } from '@/service/api'
import { useUserStore } from '@/stores'
import { theToast } from '@/utils/toast'
import { formatDate } from '@/utils/other'
import TheMdPreview from './TheMdPreview.vue'
import Verified from '../icons/verified.vue'
import MarkdownIcon from '../icons/markdown.vue'

type CapSolveDetail = {
  token?: string
}

type CapErrorDetail = {
  error?: string
}

type CapWidgetElement = HTMLElement & {
  solve?: () => Promise<CapSolveDetail>
}

const route = useRoute()
const userStore = useUserStore()
const loading = ref(false)
const submitting = ref(false)
const comments = ref<App.Api.Comment.CommentItem[]>([])
const formMeta = ref<App.Api.Comment.FormMeta | null>(null)
const captchaMountRef = ref<HTMLElement | null>(null)
const captchaWidget = ref<CapWidgetElement | null>(null)
const captchaError = ref('')
const solvingCaptcha = ref(false)
const commentFormExpanded = ref(false)
const CAP_WIDGET_SCRIPT_ID = 'cap-widget-script'

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
  return true
})

const needCaptcha = computed(() =>
  Boolean(formMeta.value?.captcha_enabled && formMeta.value?.captcha_api_endpoint),
)

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
  if (isPrivilegedUser.value) {
    return buildDiceBearURL(`${previewNickname.value}-system-preview`)
  }
  return buildDiceBearURL(`${form.nickname || 'guest'}-${form.email || 'preview'}`)
})

const resolveCommentAvatar = (item: App.Api.Comment.CommentItem, index: number) => {
  return buildDiceBearURL(`${item.id}-${item.nickname}-${item.source}-${index}`)
}

const getStickyCardStyle = (index: number) => {
  const rotateOptions = ['-0.25deg', '0deg', '0.2deg', '-0.15deg']
  const shiftOptions = ['-1px', '0px', '1px', '0px']
  return {
    '--sticky-rotate': rotateOptions[index % rotateOptions.length],
    '--sticky-shift': shiftOptions[index % shiftOptions.length],
  } as Record<string, string>
}

const clearCaptchaWidget = () => {
  if (captchaWidget.value) {
    captchaWidget.value.remove()
    captchaWidget.value = null
  }
}

const ensureCapWidgetScript = async () => {
  if (typeof window === 'undefined') return
  if (customElements.get('cap-widget')) return
  const existingScript = document.getElementById(CAP_WIDGET_SCRIPT_ID) as HTMLScriptElement | null
  if (existingScript) {
    await new Promise<void>((resolve, reject) => {
      if (customElements.get('cap-widget')) {
        resolve()
        return
      }
      existingScript.addEventListener('load', () => resolve(), { once: true })
      existingScript.addEventListener('error', () => reject(new Error('load cap widget failed')), {
        once: true,
      })
    })
    return
  }
  await new Promise<void>((resolve, reject) => {
    const script = document.createElement('script')
    script.id = CAP_WIDGET_SCRIPT_ID
    script.src = 'https://cdn.jsdelivr.net/npm/@cap.js/widget'
    script.async = true
    script.onload = () => resolve()
    script.onerror = () => reject(new Error('load cap widget failed'))
    document.head.appendChild(script)
  })
}

const mountCaptchaWidget = async () => {
  clearCaptchaWidget()
  if (!needCaptcha.value || !captchaMountRef.value) return
  try {
    await ensureCapWidgetScript()
  } catch {
    captchaError.value = '验证码组件加载失败，请稍后重试'
    return
  }
  captchaError.value = ''
  const widget = document.createElement('cap-widget') as CapWidgetElement
  widget.setAttribute('data-cap-api-endpoint', formMeta.value?.captcha_api_endpoint || '')
  widget.addEventListener('solve', (event) => {
    const token = (event as CustomEvent<CapSolveDetail>).detail?.token
    form.captcha_token = token || ''
    if (!token) {
      captchaError.value = '验证码验证失败，请重试'
      return
    }
    captchaError.value = ''
  })
  widget.addEventListener('error', (event: Event) => {
    const detail = 'detail' in event ? (event as CustomEvent<CapErrorDetail>).detail : undefined
    const message = detail?.error
    form.captcha_token = ''
    captchaError.value = message || '验证码组件异常，请刷新后重试'
  })
  captchaMountRef.value.appendChild(widget)
  captchaWidget.value = widget
}

const ensureCaptchaToken = async () => {
  if (!needCaptcha.value) return true
  if (form.captcha_token) return true
  if (!captchaWidget.value) {
    captchaError.value = '验证码组件加载中，请稍后重试'
    return false
  }
  if (!captchaWidget.value.solve) {
    captchaError.value = '请先完成验证码验证'
    return false
  }
  solvingCaptcha.value = true
  captchaError.value = ''
  try {
    const result = await captchaWidget.value.solve()
    const token = result?.token || ''
    form.captcha_token = token
    if (!token) {
      captchaError.value = '请先完成验证码验证'
      return false
    }
    return true
  } catch {
    captchaError.value = '请先完成验证码验证'
    return false
  } finally {
    solvingCaptcha.value = false
  }
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
  captchaError.value = ''
  void mountCaptchaWidget()
}

const submitComment = async () => {
  if (!canSubmit.value || submitting.value) return
  if (solvingCaptcha.value) return
  if (!(await ensureCaptchaToken())) {
    theToast.error('请先完成验证码验证')
    return
  }
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

watch(
  () => [needCaptcha.value, formMeta.value?.captcha_api_endpoint, captchaMountRef.value] as const,
  async () => {
    form.captcha_token = ''
    captchaError.value = ''
    await nextTick()
    void mountCaptchaWidget()
  },
  { immediate: true, flush: 'post' },
)

onBeforeUnmount(() => {
  clearCaptchaWidget()
})
</script>

<style scoped>
:global(:root) {
  --comment-grid-bg-color: color-mix(in srgb, var(--color-bg-canvas) 88%, #f3f2ee 12%);
  --comment-grid-line-color: rgba(120, 120, 120, 0.08);
  --comment-sticky-bg: #f8f6ee;
  --comment-sticky-border: color-mix(in srgb, var(--color-border-subtle) 78%, #d4c28f 22%);
  --comment-sticky-shadow-1: rgba(20, 20, 20, 0.05);
  --comment-sticky-shadow-2: rgba(20, 20, 20, 0.06);
}

:global(:root.dark) {
  --comment-grid-bg-color: color-mix(in srgb, var(--color-bg-canvas) 95%, #111 5%);
  --comment-grid-line-color: rgba(240, 240, 240, 0.055);
  --comment-sticky-bg: #3a3731;
  --comment-sticky-border: color-mix(in srgb, var(--color-border-subtle) 84%, #a99662 16%);
  --comment-sticky-shadow-1: rgba(0, 0, 0, 0.36);
  --comment-sticky-shadow-2: rgba(0, 0, 0, 0.32);
}

.comment-list-board {
  position: relative;
  padding: 0.75rem;
  border-radius: 10px;
}

:global(html.echo-detail-grid-bg),
:global(html.echo-detail-grid-bg body),
:global(html.echo-detail-grid-bg body #app) {
  background-color: var(--comment-grid-bg-color);
  background-image:
    linear-gradient(to right, var(--comment-grid-line-color) 1px, transparent 1px),
    linear-gradient(to bottom, var(--comment-grid-line-color) 1px, transparent 1px);
  background-size: 32px 32px;
  background-position: 0 0;
}

:global(html.echo-detail-grid-bg),
:global(html.echo-detail-grid-bg body) {
  margin: 0;
  padding: 0;
  min-height: 100%;
}

:global(html.echo-detail-grid-bg body #app) {
  min-height: 100vh;
}

.comment-sticky {
  position: relative;
  border-color: var(--comment-sticky-border);
  background: var(--comment-sticky-bg);
  box-shadow:
    0 1px 0 var(--comment-sticky-shadow-1),
    0 8px 14px var(--comment-sticky-shadow-2);
  transform: translateX(var(--sticky-shift, 0px)) rotate(var(--sticky-rotate, 0deg));
  transform-origin: 42% 8%;
  border-radius: 4px;
}

.comment-sticky::before {
  content: '';
  position: absolute;
  left: 50%;
  top: -7px;
  transform: translateX(-50%) rotate(-1.5deg);
  width: 42px;
  height: 12px;
  border-radius: 2px;
  background: color-mix(in srgb, var(--color-bg-canvas) 84%, #d7d2bf 16%);
  box-shadow:
    0 1px 0 rgba(255, 255, 255, 0.3) inset,
    0 1px 2px rgba(0, 0, 0, 0.08);
  opacity: 0.95;
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

.comment-hot-badge {
  position: absolute;
  top: 0.4rem;
  right: 0.45rem;
  z-index: 1;
  padding: 0;
  font-size: 0.67rem;
  line-height: 1.2;
  font-weight: 700;
  letter-spacing: 0.01em;
  color: #ef4444;
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

.comment-pill-btn {
  width: fit-content;
  min-width: 132px;
  max-width: 180px;
  margin-inline: auto;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.45rem;
  padding: 0.58rem 1rem;
  border-radius: 9999px;
  border: 1px solid color-mix(in srgb, var(--color-border-subtle) 82%, #cabd95 18%);
  background: color-mix(in srgb, var(--color-bg-canvas) 92%, #fff 8%);
  color: var(--color-text-primary);
  font-size: 0.9rem;
  font-weight: 600;
  transition:
    transform 0.2s ease,
    box-shadow 0.2s ease,
    border-color 0.2s ease;
  box-shadow:
    0 1px 0 rgba(20, 20, 20, 0.04),
    0 6px 12px rgba(20, 20, 20, 0.06);
}

.comment-pill-btn:hover {
  transform: translateY(-1px);
  border-color: color-mix(in srgb, var(--color-border-subtle) 72%, #b7aa7e 28%);
  box-shadow:
    0 1px 0 rgba(20, 20, 20, 0.04),
    0 10px 16px rgba(20, 20, 20, 0.09);
}

.comment-pill-btn__icon {
  width: 1.1rem;
  height: 1.1rem;
  border-radius: 9999px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 0.9rem;
  line-height: 1;
  background: color-mix(in srgb, var(--color-text-primary) 12%, transparent);
}

.comment-collapse-btn {
  padding: 0.18rem 0.55rem;
  border-radius: 9999px;
  border: 1px solid var(--color-border-subtle);
  color: var(--color-text-muted);
  font-size: 0.72rem;
  line-height: 1.2;
  transition:
    color 0.2s ease,
    border-color 0.2s ease,
    background-color 0.2s ease;
}

.comment-collapse-btn:hover {
  color: var(--color-text-primary);
  border-color: color-mix(in srgb, var(--color-border-subtle) 70%, #b7aa7e 30%);
  background: color-mix(in srgb, var(--color-bg-canvas) 85%, #fff 15%);
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

.comment-submit-row {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.75rem;
}

.comment-captcha-wrap {
  margin-right: auto;
  min-width: 0;
  max-width: 320px;
}

.comment-captcha-mount {
  min-height: 40px;
}

.comment-captcha-mount :deep(cap-widget) {
  display: block;
  max-width: 100%;
}

.comment-captcha-error {
  margin-top: 0.25rem;
}

.comment-submit-btn {
  flex-shrink: 0;
  min-height: 48px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding-inline: 1rem;
}

@media (max-width: 640px) {
  .comment-sticky {
    transform: translateX(calc(var(--sticky-shift, 0px) * 0.35))
      rotate(calc(var(--sticky-rotate, 0deg) * 0.35));
    box-shadow:
      0 1px 0 rgba(20, 20, 20, 0.06),
      0 8px 14px rgba(20, 20, 20, 0.08);
  }

  .comment-submit-row {
    flex-direction: column;
    align-items: stretch;
    gap: 0.5rem;
  }

  .comment-captcha-wrap {
    width: 100%;
    max-width: 320px;
    margin-right: 0;
    margin-inline: auto;
  }

  .comment-captcha-mount {
    display: flex;
    justify-content: center;
  }

  .comment-submit-btn {
    width: 100%;
    min-height: 40px;
  }
}
</style>
