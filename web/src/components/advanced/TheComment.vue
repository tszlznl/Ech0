<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div id="comments" class="w-full max-w-sm h-auto px-0 py-4 my-4 mx-auto">
    <div
      v-if="formMeta && !formMeta.enable_comment"
      class="rounded-lg border border-[var(--color-border-subtle)] p-3 text-sm text-[var(--color-text-muted)]"
    >
      {{ t('commentSection.disabled') }}
    </div>

    <template v-else>
      <div class="mb-4 comment-list-board">
        <div class="mb-2 flex items-center justify-between">
          <h3 class="font-semibold text-[var(--color-text-primary)]">
            {{ t('commentSection.title') }}
          </h3>
          <span class="text-xs text-[var(--color-text-muted)]">{{
            t('commentSection.count', { total: comments.length })
          }}</span>
        </div>

        <div v-if="loading" class="text-sm text-[var(--color-text-muted)]">
          {{ t('commentSection.loading') }}
        </div>
        <div v-else-if="comments.length === 0" class="text-sm text-[var(--color-text-muted)]">
          {{ t('commentSection.empty') }}
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
              <BaseAvatar
                :seed="getCommentAvatarSeed(item, index)"
                :size="32"
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
        <span>{{ t('commentSection.publishComment') }}</span>
      </button>

      <form
        v-else
        class="comment-form-panel rounded-lg border border-[var(--color-border-subtle)] p-3"
        @submit.prevent="submitComment"
      >
        <div class="mb-2 flex items-center justify-between">
          <div class="flex items-center gap-1.5">
            <h3 class="font-semibold text-[var(--color-text-primary)]">
              {{ t('commentSection.publishComment') }}
            </h3>
            <span
              class="inline-flex items-center gap-1 rounded-full px-2 py-[2px] text-[11px] text-[var(--color-text-muted)]"
              v-tooltip="t('commentSection.markdownSupported')"
            >
              <MarkdownIcon class="h-3.5 w-3.5" />
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
              :aria-label="t('commentSection.collapsePublishComment')"
              @click="commentFormExpanded = false"
            >
              {{ t('commentSection.collapse') }}
            </button>
          </div>
        </div>

        <article
          class="comment-sticky comment-preview relative mb-2 rounded-[4px] border p-3"
          :style="getStickyCardStyle(0)"
        >
          <div class="mb-2 flex items-center gap-2">
            <BaseAvatar
              :seed="previewAvatarSeed"
              :size="32"
              alt="avatar"
              class="h-8 w-8 rounded-full object-cover"
            />
            <div class="min-w-0">
              <div class="truncate text-sm font-medium text-[var(--color-text-primary)]">
                {{ previewNickname }}
              </div>
              <div class="text-xs text-[var(--color-text-muted)]">
                {{ t('commentSection.livePreview') }}
              </div>
            </div>
          </div>
          <TheMdPreview
            class="comment-md-content"
            :content="form.content || t('commentSection.previewPlaceholder')"
          />
        </article>

        <div v-if="!isPrivilegedUser" class="space-y-2">
          <input
            v-model.trim="form.nickname"
            type="text"
            class="comment-input-field w-full rounded-md border border-[var(--color-border-subtle)] bg-transparent px-3 py-2 text-sm"
            :placeholder="t('commentSection.nicknameRequired')"
          />
          <input
            v-model.trim="form.email"
            type="email"
            class="comment-input-field w-full rounded-md border border-[var(--color-border-subtle)] bg-transparent px-3 py-2 text-sm"
            :placeholder="t('commentSection.emailRequired')"
          />
          <input
            v-model.trim="form.website"
            type="url"
            class="comment-input-field w-full rounded-md border border-[var(--color-border-subtle)] bg-transparent px-3 py-2 text-sm"
            :placeholder="t('commentSection.websiteOptional')"
          />
        </div>

        <textarea
          v-model.trim="form.content"
          class="comment-input-field comment-textarea mt-2 min-h-24 w-full rounded-md border border-[var(--color-border-subtle)] bg-transparent px-3 py-2 text-sm"
          :placeholder="t('commentSection.commentPlaceholder')"
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
        />

        <div class="comment-submit-row mt-3">
          <div v-if="needCaptcha" class="comment-captcha-wrap">
            <div ref="captchaMountRef" class="comment-captcha-mount"></div>
            <p v-if="captchaError" class="comment-captcha-error text-xs text-red-500">
              {{ captchaError }}
            </p>
          </div>
          <button
            v-if="showSubmitButton"
            type="submit"
            class="comment-submit-btn rounded-md bg-[var(--color-text-primary)] px-4 py-1 text-sm text-[var(--color-bg-canvas)]"
            :disabled="submitting || !canSubmit"
          >
            {{ submitting ? t('commentSection.submitting') : t('commentSection.submitComment') }}
          </button>
        </div>

        <div
          v-if="submitNotice"
          class="mt-3 rounded-md border border-[var(--color-border-subtle)] bg-[var(--color-bg-canvas)] p-3 text-sm"
        >
          <div class="font-medium text-[var(--color-text-primary)]">
            {{
              submitNotice.status === 'approved'
                ? t('commentSection.commentPublished')
                : t('commentSection.commentSubmittedPending')
            }}
          </div>
          <div class="mt-1 text-xs text-[var(--color-text-muted)]">
            {{ submitNoticeText }}
          </div>
          <div
            v-if="submitNotice.contentPreview"
            class="mt-2 rounded border border-[var(--color-border-subtle)] px-2 py-1 text-xs text-[var(--color-text-muted)]"
          >
            {{ submitNotice.contentPreview }}
          </div>
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
import BaseAvatar from '@/components/common/BaseAvatar.vue'
import { TheMdPreview } from '@/components/advanced/md'
import Verified from '../icons/verified.vue'
import MarkdownIcon from '../icons/markdown.vue'
import { useI18n } from 'vue-i18n'

type CapSolveDetail = {
  token?: string
}

type CapErrorDetail = {
  error?: string
  message?: string
}

type CapWidgetElement = HTMLElement & {
  solve?: () => Promise<CapSolveDetail>
}

type SubmitNotice = {
  status: App.Api.Comment.CommentStatus
  contentPreview: string
  submittedAt: number
}

const route = useRoute()
const userStore = useUserStore()
const { t, locale } = useI18n()
const loading = ref(false)
const submitting = ref(false)
const comments = ref<App.Api.Comment.CommentItem[]>([])
const formMeta = ref<App.Api.Comment.FormMeta | null>(null)
const captchaMountRef = ref<HTMLElement | null>(null)
const captchaWidget = ref<CapWidgetElement | null>(null)
const captchaError = ref('')
const solvingCaptcha = ref(false)
const commentFormExpanded = ref(false)
const submitNotice = ref<SubmitNotice | null>(null)
let capWidgetLoadPromise: Promise<unknown> | null = null

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
const showSubmitButton = computed(() => !needCaptcha.value || Boolean(form.captcha_token))

const contentLength = computed(() => form.content.length)
const contentTooLong = computed(() => contentLength.value > 200)

const profileReady = computed(() => {
  if (isPrivilegedUser.value) return true
  return Boolean(form.nickname && form.email)
})

const previewNickname = computed(() => {
  if (isPrivilegedUser.value) return userStore.user?.username || String(t('commentSection.you'))
  return form.nickname || String(t('commentSection.you'))
})

const previewAvatarSeed = computed(() => {
  if (isPrivilegedUser.value) {
    return `${previewNickname.value}-system-preview`
  }
  return `${form.nickname || 'guest'}-${form.email || 'preview'}`
})

const getCommentAvatarSeed = (item: App.Api.Comment.CommentItem, index: number) => {
  return `${item.id}-${item.nickname}-${item.source}-${index}`
}

const getStickyCardStyle = (index: number) => {
  const rotateOptions = ['-0.25deg', '0deg', '0.2deg', '-0.15deg']
  const shiftOptions = ['-1px', '0px', '1px', '0px']
  return {
    '--sticky-rotate': rotateOptions[index % rotateOptions.length],
    '--sticky-shift': shiftOptions[index % shiftOptions.length],
  } as Record<string, string>
}

const submitNoticeText = computed(() => {
  if (!submitNotice.value) return ''
  const timeLabel = formatDate(submitNotice.value.submittedAt)
  if (submitNotice.value.status === 'approved') {
    return String(t('commentSection.publishedAt', { time: timeLabel }))
  }
  return String(t('commentSection.submittedAtPending', { time: timeLabel }))
})

const clearCaptchaWidget = () => {
  if (captchaWidget.value) {
    captchaWidget.value.remove()
    captchaWidget.value = null
  }
}

const ensureCapWidgetScript = async () => {
  if (typeof window === 'undefined') return
  if (customElements.get('cap-widget')) return
  capWidgetLoadPromise ??= import('@cap.js/widget')
  await capWidgetLoadPromise
}

const mountCaptchaWidget = async () => {
  clearCaptchaWidget()
  if (!needCaptcha.value || !captchaMountRef.value) return
  try {
    await ensureCapWidgetScript()
  } catch {
    captchaError.value = String(t('commentSection.captchaLoadFailed'))
    return
  }
  captchaError.value = ''
  const widget = document.createElement('cap-widget') as CapWidgetElement
  widget.setAttribute('data-cap-api-endpoint', formMeta.value?.captcha_api_endpoint || '')
  const i18nAttrs: Record<string, string> = {
    'data-cap-i18n-initial-state': String(t('commentSection.capInitialState')),
    'data-cap-i18n-verifying-label': String(t('commentSection.capVerifyingLabel')),
    'data-cap-i18n-solved-label': String(t('commentSection.capSolvedLabel')),
    'data-cap-i18n-error-label': String(t('commentSection.capErrorLabel')),
    'data-cap-i18n-troubleshooting-label': String(t('commentSection.capTroubleshootingLabel')),
    'data-cap-i18n-wasm-disabled': String(t('commentSection.capWasmDisabled')),
    'data-cap-i18n-verify-aria-label': String(t('commentSection.capVerifyAriaLabel')),
    'data-cap-i18n-verifying-aria-label': String(t('commentSection.capVerifyingAriaLabel')),
    'data-cap-i18n-verified-aria-label': String(t('commentSection.capVerifiedAriaLabel')),
    'data-cap-i18n-error-aria-label': String(t('commentSection.capErrorAriaLabel')),
  }
  Object.entries(i18nAttrs).forEach(([key, value]) => widget.setAttribute(key, value))
  widget.addEventListener('solve', (event) => {
    const token = (event as CustomEvent<CapSolveDetail>).detail?.token
    form.captcha_token = token || ''
    if (!token) {
      captchaError.value = String(t('commentSection.captchaVerifyFailed'))
      return
    }
    captchaError.value = ''
  })
  widget.addEventListener('error', (event: Event) => {
    const detail = 'detail' in event ? (event as CustomEvent<CapErrorDetail>).detail : undefined
    const message = detail?.message || detail?.error
    form.captcha_token = ''
    captchaError.value = message || String(t('commentSection.captchaWidgetError'))
  })
  captchaMountRef.value.appendChild(widget)
  captchaWidget.value = widget
}

const ensureCaptchaToken = async () => {
  if (!needCaptcha.value) return true
  if (form.captcha_token) return true
  if (!captchaWidget.value) {
    captchaError.value = String(t('commentSection.captchaLoading'))
    return false
  }
  if (!captchaWidget.value.solve) {
    captchaError.value = String(t('commentSection.completeCaptchaFirst'))
    return false
  }
  solvingCaptcha.value = true
  captchaError.value = ''
  try {
    const result = await captchaWidget.value.solve()
    const token = result?.token || ''
    form.captcha_token = token
    if (!token) {
      captchaError.value = String(t('commentSection.completeCaptchaFirst'))
      return false
    }
    return true
  } catch {
    captchaError.value = String(t('commentSection.completeCaptchaFirst'))
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
    theToast.error(String(t('commentSection.completeCaptchaFirst')))
    return
  }
  submitting.value = true
  try {
    const submittedContent = form.content.trim()
    const res = await fetchCreateComment(form)
    if (res.code === 1) {
      const status = res.data?.status || 'pending'
      submitNotice.value = {
        status,
        contentPreview: submittedContent.slice(0, 80),
        submittedAt: Date.now(),
      }
      if (status === 'approved') {
        theToast.success(String(t('commentSection.commentPublished')))
      } else {
        theToast.success(String(t('commentSection.commentSubmittedPending')))
      }
      resetForm()
      if (status === 'approved') {
        await loadData()
      }
    }
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  loadData()
})

watch(
  () =>
    [
      needCaptcha.value,
      formMeta.value?.captcha_api_endpoint,
      captchaMountRef.value,
      locale.value,
    ] as const,
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
  --comment-grid-bg-color: var(--comment-grid-bg);
  --comment-grid-line-color: rgb(120 120 120 / 8%);
  --comment-sticky-bg: #f8f6ee;
  --comment-sticky-border: var(--comment-sticky-border-color);
  --comment-sticky-shadow-1: rgb(20 20 20 / 5%);
  --comment-sticky-shadow-2: rgb(20 20 20 / 6%);
}

:global(:root.dark) {
  --comment-grid-bg-color: var(--comment-grid-bg);
  --comment-grid-line-color: rgb(240 240 240 / 5.5%);
  --comment-sticky-bg: #3a3731;
  --comment-sticky-border: var(--comment-sticky-border-color);
  --comment-sticky-shadow-1: rgb(0 0 0 / 36%);
  --comment-sticky-shadow-2: rgb(0 0 0 / 32%);
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
  transform: translateX(var(--sticky-shift, 0)) rotate(var(--sticky-rotate, 0deg));
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
  background: var(--comment-sticky-before-bg);
  box-shadow:
    0 1px 0 rgb(255 255 255 / 30%) inset,
    0 1px 2px rgb(0 0 0 / 8%);
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
  box-shadow: 0 0 0 1px rgb(255 255 255 / 72%);
}

.comment-ready-dot.is-ready {
  background: #10b981;
}

.comment-ready-dot.is-not-ready {
  background: #ef4444;
}

.comment-form-panel {
  background: var(--comment-form-panel-bg);
  box-shadow:
    0 1px 0 rgb(20 20 20 / 4%),
    0 10px 18px rgb(20 20 20 / 8%);
  border-color: var(--comment-form-panel-border);
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
  border: 1px solid var(--comment-pill-btn-border);
  background: var(--comment-form-panel-bg);
  color: var(--color-text-primary);
  font-size: 0.9rem;
  font-weight: 600;
  transition:
    transform 0.2s ease,
    box-shadow 0.2s ease,
    border-color 0.2s ease;
  box-shadow:
    0 1px 0 rgb(20 20 20 / 4%),
    0 6px 12px rgb(20 20 20 / 6%);
}

.comment-pill-btn:hover {
  transform: translateY(-1px);
  border-color: var(--comment-pill-btn-hover-border);
  box-shadow:
    0 1px 0 rgb(20 20 20 / 4%),
    0 10px 16px rgb(20 20 20 / 9%);
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
  background: var(--comment-pill-icon-bg);
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
  border-color: var(--comment-collapse-hover-border);
  background: var(--comment-collapse-hover-bg);
}

.comment-md-content {
  color: var(--color-text-primary);
  font-size: 0.9rem;
  line-height: 1.65;
}

.comment-md-content :deep(p) {
  margin: 0.15rem 0;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}

.comment-preview {
  opacity: 0.98;
}

.comment-submit-row {
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 0.75rem;
}

.comment-captcha-wrap {
  width: 100%;
  box-sizing: border-box;
}

.comment-captcha-mount {
  min-height: 48px;
}

.comment-captcha-mount :deep(cap-widget) {
  box-sizing: border-box;
  display: block;
  width: 100% !important;
  max-width: 100% !important;
  min-width: 0;

  --cap-widget-width: 100%;
}

.comment-captcha-error {
  margin-top: 0.25rem;
}

.comment-submit-btn {
  box-sizing: border-box;
  width: 100%;
  min-height: 48px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding-inline: 1rem;
  white-space: nowrap;
}

.comment-input-field {
  color: var(--color-text-primary);
  border-color: var(--comment-input-border);
  background: var(--comment-input-bg);
  transition:
    border-color 0.2s ease,
    box-shadow 0.22s ease,
    background-color 0.2s ease;
}

.comment-input-field::placeholder {
  color: var(--comment-input-placeholder);
}

.comment-input-field:hover {
  border-color: var(--comment-input-hover-border);
}

.comment-input-field:focus,
.comment-input-field:focus-visible {
  outline: none;
  border-color: var(--comment-input-focus-border);
  background: var(--comment-input-focus-bg);
  box-shadow:
    0 0 0 1px var(--comment-input-focus-ring-inner),
    0 0 0 4px var(--comment-input-focus-ring-outer);
}

.comment-textarea {
  line-height: 1.6;
  resize: vertical;
}

@media (width <= 640px) {
  .comment-sticky {
    transform: translateX(calc(var(--sticky-shift, 0px) * 0.35))
      rotate(calc(var(--sticky-rotate, 0deg) * 0.35));
    box-shadow:
      0 1px 0 rgb(20 20 20 / 6%),
      0 8px 14px rgb(20 20 20 / 8%);
  }

  .comment-submit-row {
    display: flex;
    flex-direction: column;
    align-items: stretch;
    gap: 0.5rem;
  }

  .comment-captcha-wrap {
    flex: 1 1 auto;
    width: 100%;
    margin-right: 0;
  }

  .comment-captcha-mount {
    min-height: 40px;
  }

  .comment-captcha-mount :deep(cap-widget) {
    width: 100% !important;
    max-width: 100% !important;

    --cap-widget-height: 40px;
  }

  .comment-submit-btn {
    width: 100%;
    margin-left: 0;
    min-height: 40px;
  }
}
</style>
