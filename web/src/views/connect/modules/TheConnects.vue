<template>
  <div class="px-9 md:px-11">
    <!-- 列出所有连接（列出每个连接的头像） -->
    <div class="widget bg-transparent! w-full max-w-[19rem] mx-auto rounded-md p-4">
      <div class="connect-head mb-2">
        <div class="connect-icon-chip">
          <Connect class="w-8 h-8" />
        </div>
        <div class="connect-title-wrap">
          <div class="connect-title">Connect</div>
          <div class="connect-title-accent">Widget</div>
        </div>
      </div>
      <div v-if="!loading">
        <div v-if="!connectsInfo.length" class="text-[var(--color-text-muted)] text-sm mb-2">
          当前暂无连接
        </div>
        <div v-else class="flex flex-wrap gap-3">
          <div
            v-for="(connect, index) in connectsInfo"
            :key="index"
            class="relative flex flex-col items-center justify-center w-8 h-8 min-w-[2rem] min-h-[2rem] flex-none border-2 border-[var(--color-border-subtle)] shadow-sm rounded-full hover:shadow-md transition duration-200 ease-in-out group"
          >
            <a :href="connect.server_url" target="_blank" class="block w-full h-full">
              <img
                :src="connect.logo"
                alt="avatar"
                class="w-full h-full rounded-full object-cover"
              />
              <!-- 热力圆点 -->
              <span
                class="absolute top-0 right-0 w-2.5 h-2.5 border-2 border-[var(--color-bg-surface)] rounded-full"
                :style="{
                  transform: 'translate(35%, -35%)',
                  backgroundColor: getColor(connect.today_echos || 0),
                }"
              ></span>
            </a>
            <!-- Tooltip -->
            <div
              class="absolute z-10 left-1/2 -translate-x-1/2 top-10 min-w-max bg-gray-800 text-white text-xs rounded px-3 py-2 opacity-0 group-hover:opacity-100 pointer-events-none transition-opacity duration-200 shadow-lg"
            >
              <div class="font-bold mb-1">{{ connect.server_name }}</div>
              <div>Owner: {{ connect.sys_username || '-' }}</div>
              <div>共有: {{ connect.total_echos ?? 0 }}</div>
              <div>今日: {{ connect.today_echos ?? 0 }}</div>
              <div>版本: {{ connect.version || '-' }}</div>
            </div>
          </div>
        </div>
      </div>
      <div v-else>
        <div class="text-[var(--color-text-secondary)] text-sm mb-2">加载中...</div>
      </div>

      <div class="comment-teaser mt-8">
        <div class="comment-head mb-2">
          <div class="comment-icon-chip">
            <div class="doodle-smile" aria-hidden="true">
              <span class="smile-eye smile-eye-left"></span>
              <span class="smile-eye smile-eye-right"></span>
              <span class="smile-mouth"></span>
            </div>
          </div>
          <div class="comment-title-wrap">
            <div class="comment-title">Comment</div>
            <div class="comment-title-accent">Random</div>
          </div>
        </div>
        <div class="comment-teaser-body">
          <div class="comment-teaser-card">
            <button
              v-if="canJumpToEchoDetail"
              type="button"
              class="comment-jump-btn"
              title="查看对应 Echo 详情"
              aria-label="查看对应 Echo 详情"
              @click="handleJumpToEchoDetail"
            >
              <LinkTo class="w-4 h-4" />
            </button>
            <p v-if="commentLoading" class="comment-teaser-content">正在挑选一条精选评论...</p>
            <p v-else class="comment-teaser-content">{{ randomCommentContent }}</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Connect from '@/components/icons/connect.vue'
import LinkTo from '@/components/icons/linkto.vue'
import { fetchGetPublicComments } from '@/service/api'
import { useConnectStore } from '@/stores'
import { storeToRefs } from 'pinia'
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

const connectStore = useConnectStore()
const router = useRouter()
const { getConnectInfo } = connectStore
const { loading, connectsInfo } = storeToRefs(connectStore)
const randomComment = ref<App.Api.Comment.CommentItem | null>(null)
const commentLoading = ref(false)
const canJumpToEchoDetail = computed(() => Boolean(randomComment.value?.echo_id) && !commentLoading.value)

const randomCommentContent = computed(() => {
  const content = randomComment.value?.content?.trim()
  if (!content) return '暂无精选评论可展示。'
  return content.replace(/\s+/g, ' ').slice(0, 120)
})

const getColor = (count: number): string => {
  if (count >= 4) return 'var(--color-accent)'
  if (count >= 3) return 'var(--color-accent)'
  if (count >= 2) return 'var(--color-accent)'
  if (count >= 1) return 'var(--color-accent-soft)'
  return '#c4c3c1'
}

const pickRandomComment = (items: App.Api.Comment.CommentItem[]) => {
  const hotComments = items.filter(
    (item) => item.status === 'approved' && item.hot && item.content?.trim(),
  )
  if (hotComments.length > 0) {
    randomComment.value = hotComments[Math.floor(Math.random() * hotComments.length)]
    return
  }

  const approvedComments = items.filter(
    (item) => item.status === 'approved' && item.content?.trim(),
  )
  if (approvedComments.length > 0) {
    randomComment.value = approvedComments[Math.floor(Math.random() * approvedComments.length)]
  }
}

const loadRandomComment = async () => {
  commentLoading.value = true
  try {
    const res = await fetchGetPublicComments(30)
    if (res.code === 1 && res.data?.length) {
      pickRandomComment(res.data)
    }
  } catch {
    randomComment.value = null
  } finally {
    commentLoading.value = false
  }
}

const handleJumpToEchoDetail = () => {
  const echoId = randomComment.value?.echo_id?.trim()
  if (!echoId) return
  router.push({
    name: 'echo',
    params: { echoId },
  })
}

onMounted(() => {
  getConnectInfo()
  void loadRandomComment()
})
</script>

<style scoped>
.connect-head {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 12px;
}

.connect-icon-chip {
  width: 64px;
  height: 64px;
  border-radius: 9999px;
  color: var(--color-text-muted);
  display: flex;
  align-items: center;
  justify-content: center;
}

.connect-title-wrap {
  line-height: 0.9;
  text-align: right;
}

.connect-title {
  font-family: Georgia, 'Times New Roman', serif;
  font-size: 26px;
  font-weight: 600;
  color: var(--color-text-primary);
}

.connect-title-accent {
  font-family: var(--font-family-handwritten);
  color: var(--color-accent);
  font-size: 18px;
  margin-top: -2px;
}

.comment-head {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 12px;
}

.comment-icon-chip {
  width: 52px;
  height: 52px;
  border-radius: 9999px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.doodle-smile {
  position: relative;
  width: 34px;
  height: 34px;
  border-radius: 9999px;
  background: #f2e68f;
  border: 1.5px solid #3f362d;
  transform: rotate(-5deg);
}

.smile-eye {
  position: absolute;
  width: 2px;
  height: 2px;
  border-radius: 9999px;
  background: #3f362d;
  top: 13px;
}

.smile-eye-left {
  left: 10px;
}

.smile-eye-right {
  left: 19px;
}

.smile-mouth {
  position: absolute;
  left: 11px;
  top: 18px;
  width: 10px;
  height: 6px;
  border-bottom: 1.5px solid #3f362d;
  border-radius: 0 0 10px 10px;
}

.comment-title-wrap {
  line-height: 0.9;
  text-align: right;
}

.comment-title {
  font-family: Georgia, 'Times New Roman', serif;
  color: var(--color-text-primary);
  font-size: 26px;
  font-weight: 600;
}

.comment-title-accent {
  font-family: var(--font-family-handwritten);
  color: var(--color-accent);
  font-size: 18px;
  margin-top: -2px;
}

.comment-teaser-body {
  width: 100%;
}

.comment-teaser-card {
  position: relative;
  width: 100%;
  border: 1px solid var(--color-border-subtle);
  background-color: color-mix(in srgb, var(--color-bg-surface) 78%, transparent);
  box-shadow: 0 8px 18px rgba(20, 20, 20, 0.04);
  padding: 14px 12px 12px;
  transform: rotate(-1.1deg);
  transform-origin: top center;
  transition: transform 220ms ease;
}

.comment-teaser-card:hover {
  transform: rotate(-0.4deg);
}

.comment-jump-btn {
  position: absolute;
  right: 8px;
  top: 8px;
  width: 26px;
  height: 26px;
  border-radius: 9999px;
  border: 1px solid color-mix(in srgb, var(--color-border-subtle) 88%, transparent);
  background: color-mix(in srgb, var(--color-bg-canvas) 92%, transparent);
  color: var(--color-text-secondary);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transform: translateY(-1px) scale(0.96);
  pointer-events: none;
  transition:
    opacity 180ms ease,
    transform 180ms ease,
    color 180ms ease,
    border-color 180ms ease;
}

.comment-teaser-card:hover .comment-jump-btn,
.comment-jump-btn:focus-visible {
  opacity: 1;
  transform: translateY(0) scale(1);
  pointer-events: auto;
}

.comment-jump-btn:hover {
  color: var(--color-text-primary);
  border-color: var(--color-accent);
}

.comment-teaser-card::before {
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

.comment-teaser-content {
  color: var(--color-text-secondary);
  font-size: 13px;
  line-height: 1.65;
  white-space: normal;
  word-break: break-word;
}
</style>
