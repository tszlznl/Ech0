<template>
  <div class="panel-page-wrap mt-4">
    <div class="panel-shell border p-4 mx-auto flex flex-col max-w-screen-lg w-full mb-12">
      <section class="panel-welcome mb-5 md:mb-2">
        <div class="panel-welcome-main">
          <h1 class="panel-welcome-username">
            {{ username }}，欢迎回来 <span class="wave-hand">👋</span>
          </h1>
          <div class="panel-welcome-meta">
            <p class="panel-welcome-date">{{ dateText }}</p>
            <p class="panel-welcome-greeting">{{ greeting }}</p>
          </div>
          <p class="panel-welcome-tip">今天也来记录一点新的灵感</p>
        </div>
        <div class="panel-welcome-actions">
          <BaseButton
            :icon="BackHand"
            @click="router.push('/')"
            class="panel-home-btn"
            title="返回首页"
          />
        </div>
      </section>

      <!-- 移动端选择器 -->
      <div class="md:hidden mb-6 px-2 flex justify-between items-center mb-3">
        <div class="w-1/2">
          <BaseSelect
            class="!focus:ring-0 h-9"
            v-model="selectedRoute"
            :options="routeOptions"
            placeholder="选择页面"
            @change="handleRouteChange"
          />
        </div>

        <div class="flex gap-2 items-center">
          <!-- 退出登录 -->
          <BaseButton
            v-if="userStore.isLogin"
            :icon="Logout"
            @click="handleLogout"
            class="w-9 h-9 rounded-md"
            title="退出"
          >
          </BaseButton>
          <!-- 登录 / 注册 -->
          <BaseButton
            v-else
            :icon="Auth"
            @click="router.push('/auth')"
            class="w-9 h-9 rounded-md"
            title="登录 / 注册"
          >
          </BaseButton>
        </div>
      </div>

      <!-- 主内容区 -->
      <div class="mx-auto flex my-4 w-full max-w-screen-lg rounded-md panel-main">
        <!-- 桌面端侧边栏 -->
        <div class="hidden md:flex flex-col gap-2 w-48 pr-8 shrink-0 panel-nav">
          <!-- Dashboard -->
          <BaseButton
            :icon="Dashboard"
            @click="router.push('/panel/dashboard')"
            :class="getButtonClasses('panel-dashboard')"
            title="Dashboard"
          >
            Dashboard
          </BaseButton>

          <!-- 偏好设置 -->
          <BaseButton
            :icon="Setting"
            @click="router.push('/panel/setting')"
            :class="getButtonClasses('panel-setting')"
            title="偏好设置"
          >
            偏好设置
          </BaseButton>

          <!-- 用户中心 -->
          <BaseButton
            :icon="User"
            @click="router.push('/panel/user')"
            :class="getButtonClasses('panel-user')"
            title="用户中心"
          >
            用户中心
          </BaseButton>

          <!-- 存储管理 -->
          <BaseButton
            :icon="Storage"
            @click="router.push('/panel/storage')"
            :class="getButtonClasses('panel-storage')"
            title="存储管理"
          >
            存储管理
          </BaseButton>

          <!-- 数据管理 -->
          <BaseButton
            :icon="Data"
            @click="router.push('/panel/data-management')"
            :class="getButtonClasses('panel-data-management')"
            title="数据管理"
          >
            数据管理
          </BaseButton>

          <!-- 单点登录 -->
          <BaseButton
            :icon="Sso"
            @click="router.push('/panel/sso')"
            :class="getButtonClasses('panel-sso')"
            title="单点登录"
          >
            单点登录
          </BaseButton>

          <!-- 功能扩展 -->
          <BaseButton
            :icon="Extension"
            @click="router.push('/panel/extension')"
            :class="getButtonClasses('panel-extension')"
            title="功能扩展"
          >
            功能扩展
          </BaseButton>

          <!-- 外部集成 -->
          <BaseButton
            :icon="Others"
            @click="router.push('/panel/advance')"
            :class="getButtonClasses('panel-advance')"
            title="外部集成"
          >
            外部集成
          </BaseButton>

          <!-- 系统日志 -->
          <BaseButton
            :icon="Log"
            @click="router.push('/panel/system-log')"
            :class="getButtonClasses('panel-system-log')"
            title="系统日志"
          >
            系统日志
          </BaseButton>

          <div class="h-px bg-[var(--color-border-subtle)] mx-2" />

          <!-- 退出登录 -->
          <BaseButton
            :icon="Logout"
            @click="handleLogout"
            :class="getBottomButtonClasses()"
            title="退出登录"
          >
            登出
          </BaseButton>

          <!-- 登录 / 注册 -->
          <BaseButton
            :icon="Auth"
            @click="router.push('/auth')"
            :class="getBottomButtonClasses()"
            title="登录 / 注册"
          >
            登录
          </BaseButton>

          <div class="panel-version my-2 ml-3">Version: {{ settingStore.hello?.version }}</div>
        </div>

        <!-- 路由内容 -->
        <div class="flex-1 min-w-0 panel-content">
          <router-view />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import BaseButton from '@/components/common/BaseButton.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import User from '@/components/icons/user.vue'
import Auth from '@/components/icons/auth.vue'
import BackHand from '@/components/icons/backhand.vue'
import Extension from '@/components/icons/extension.vue'
import Others from '@/components/icons/theothers.vue'
import Dashboard from '@/components/icons/dashboard.vue'
import Setting from '@/components/icons/setting.vue'
import Storage from '@/components/icons/storage.vue'
import Data from '@/components/icons/data.vue'
import Sso from '@/components/icons/sso.vue'
import Logout from '@/components/icons/logout.vue'
import Log from '@/components/icons/log.vue'
import { computed, ref, watch } from 'vue'
import { useUserStore, useSettingStore } from '@/stores'
import { useRouter, useRoute } from 'vue-router'
import { theToast } from '@/utils/toast'
import { useBaseDialog } from '@/composables/useBaseDialog'

const { openConfirm } = useBaseDialog()

const userStore = useUserStore()
const settingStore = useSettingStore()
const router = useRouter()
const route = useRoute()

const currentRoute = computed(() => route.name as string)
const selectedRoute = ref(route.path)
const username = computed(() => userStore.user?.username || '朋友')
const greeting = computed(() => {
  const hour = new Date().getHours()
  if (hour < 6) return '凌晨好'
  if (hour < 12) return '早上好'
  if (hour < 18) return '下午好'
  return '晚上好'
})
const dateText = computed(() => {
  return new Intl.DateTimeFormat('zh-CN', {
    month: 'long',
    day: 'numeric',
    weekday: 'long',
  }).format(new Date())
})

// 统一的按钮样式计算函数
const getButtonClasses = (routeName: string) => {
  const baseClasses =
    'flex items-center gap-2 pl-3 py-1.5 rounded-[var(--radius-md)] transition-all duration-200 border-none !shadow-none !ring-0 justify-start bg-transparent hover:bg-[var(--color-bg-muted)]'
  const activeClasses =
    currentRoute.value === routeName
      ? 'text-[var(--color-nav-active-text)]! bg-[var(--color-nav-active-bg)]!'
      : 'text-[var(--color-text-secondary)]'

  return `${baseClasses} ${activeClasses}`
}

// 底部按钮样式
const getBottomButtonClasses = () => {
  return 'flex items-center gap-2 pl-3 py-1.5 rounded-[var(--radius-md)] transition-all duration-200 border-none !shadow-none !ring-0 text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-muted)] justify-start bg-transparent'
}

// 路由选项
const routeOptions = [
  { label: 'Dashboard', value: '/panel/dashboard' },
  { label: '偏好设置', value: '/panel/setting' },
  { label: '用户中心', value: '/panel/user' },
  { label: '存储管理', value: '/panel/storage' },
  { label: '数据管理', value: '/panel/data-management' },
  { label: '单点登录', value: '/panel/sso' },
  { label: '功能扩展', value: '/panel/extension' },
  { label: '外部集成', value: '/panel/advance' },
  { label: '系统日志', value: '/panel/system-log' },
]

// 监听路由变化，更新选择器
watch(
  () => route.path,
  (newPath) => {
    selectedRoute.value = newPath
  },
)

// 处理选择器变化
const handleRouteChange = () => {
  router.push(selectedRoute.value)
}

const handleLogout = () => {
  // 检查是否登录
  if (!userStore.isLogin) {
    theToast.info('当前未登录')
    return
  }

  // 弹出浏览器确认框
  openConfirm({
    title: '确定要退出登录吗？',
    description: '',
    onConfirm: () => {
      // 清除用户信息
      userStore.logout()
      // 跳转到首页
      router.push('/')
      theToast.success('已退出登录')
    },
  })
}
</script>

<style scoped>
.panel-page-wrap {
  min-height: calc(100vh - 3.5rem);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0.6rem 0.5rem 0;
}

.panel-shell {
  border-color: var(--color-border-subtle);
  border-radius: var(--radius-lg);
  background: var(--color-bg-canvas);
}

.panel-nav {
  color: var(--color-text-secondary);
}

.panel-main {
  align-items: flex-start;
  gap: 0.25rem;
}

.panel-content {
  max-width: 53rem;
  width: 100%;
  padding-left: 0.2rem;
}

.panel-version {
  color: var(--color-text-muted);
  font-family: var(--font-family-display);
}

.panel-welcome {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 0.5rem 0.25rem 0.65rem;
  border-bottom: 1px solid var(--color-border-subtle);
}

.panel-welcome-main {
  min-width: 0;
  text-align: left;
}

.panel-welcome-username {
  margin: 0;
  font-size: clamp(1.45rem, 2.4vw, 1.9rem);
  line-height: 1.2;
  color: var(--color-text-primary);
  font-weight: 800;
  font-family: var(--font-family-display);
  letter-spacing: 0.01em;
}

.panel-welcome-greeting {
  font-size: 0.9rem;
  line-height: 1.2;
  color: var(--color-text-muted);
  font-weight: 600;
}

.panel-welcome-meta {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  margin-top: 0.35rem;
}

.panel-welcome-tip {
  margin-top: 0.3rem;
  color: var(--color-text-muted);
  font-size: 0.88rem;
}

.panel-welcome-date {
  display: inline-flex;
  align-items: center;
  color: var(--color-text-secondary);
  font-size: 0.9rem;
  white-space: nowrap;
  font-weight: 600;
}

.panel-welcome-actions {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 0.5rem;
}

.panel-home-btn {
  border: 1px solid var(--color-border-subtle) !important;
  border-radius: 999px !important;
  background: var(--color-bg-surface) !important;
  color: var(--color-text-secondary) !important;
  width: 2rem !important;
  height: 2rem !important;
  padding: 0 !important;
  display: inline-flex !important;
  align-items: center !important;
  justify-content: center !important;
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease;
}

.panel-home-btn:hover {
  border-color: var(--color-border-strong) !important;
  background: var(--color-bg-muted) !important;
}

.wave-hand {
  display: inline-block;
  margin-left: 0.2rem;
  transform-origin: center;
  will-change: transform;
}

.wave-hand:hover {
  animation: hand-shake 620ms ease-in-out;
}

@keyframes hand-shake {
  0% {
    transform: rotate(0deg) scale(1);
  }
  15% {
    transform: rotate(16deg) scale(1.08);
  }
  30% {
    transform: rotate(-14deg) scale(1.08);
  }
  45% {
    transform: rotate(12deg) scale(1.06);
  }
  60% {
    transform: rotate(-10deg) scale(1.04);
  }
  75% {
    transform: rotate(7deg) scale(1.02);
  }
  100% {
    transform: rotate(0deg) scale(1);
  }
}

@media (max-width: 768px) {
  .panel-page-wrap {
    min-height: auto;
    display: block;
    padding: 0.4rem 0.35rem 0;
  }

  .panel-welcome {
    align-items: flex-start;
    flex-direction: row;
    justify-content: space-between;
    gap: 0.6rem;
  }

  .panel-welcome-date {
    white-space: normal;
  }

  .panel-main {
    gap: 0;
  }

  .panel-content {
    max-width: 100%;
    padding-left: 0;
  }
}
</style>
