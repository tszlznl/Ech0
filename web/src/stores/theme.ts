import { defineStore } from 'pinia'
import { ref } from 'vue'
import { localStg } from '@/utils/storage'

export type ThemeMode = 'light' | 'dark' | 'sunny'
type ThemeType = 'light' | 'dark' | 'sunny'
const THEME_COLOR_META_NAME = 'theme-color'
const THEME_COLOR_FALLBACK: Record<ThemeType, string> = {
  light: '#f4f1ec',
  dark: '#333333',
  sunny: 'rgb(238, 236, 230)', // 需要考虑有一层视频带来的色彩层级变化 RGB(238,236,230)
}

export const useThemeStore = defineStore('themeStore', () => {
  const savedThemeMode = localStg.getItem('themeMode')
  const savedTheme = localStg.getItem('theme')

  // 初始化 themeMode
  const mode = ref<ThemeMode>(
    savedThemeMode === 'light' || savedThemeMode === 'dark' || savedThemeMode === 'sunny'
      ? savedThemeMode
      : 'light',
  )
  const theme = ref<ThemeType>(
    savedTheme === 'light' || savedTheme === 'dark' || savedTheme === 'sunny'
      ? savedTheme
      : 'light',
  )

  // 内部切换主题逻辑
  const applyThemeToggle = () => {
    if (mode.value === 'light') {
      mode.value = 'sunny'
    } else if (mode.value === 'sunny') {
      mode.value = 'dark'
    } else {
      mode.value = 'light'
    }

    applyTheme()
    localStg.setItem('themeMode', mode.value)
  }

  // 防抖标志：防止动画过程中重复触发
  let isTransitioning = false

  // 带扩散动画的主题切换
  const toggleTheme = async (event?: MouseEvent) => {
    // 防抖：如果正在过渡中，忽略此次点击
    if (isTransitioning) return

    // 获取点击坐标，如果没有事件则从屏幕中心扩散
    const x = event?.clientX ?? window.innerWidth / 2
    const y = event?.clientY ?? window.innerHeight / 2

    // 计算到最远角的距离（用于确定圆形大小）
    const endRadius = Math.hypot(
      Math.max(x, window.innerWidth - x),
      Math.max(y, window.innerHeight - y),
    )

    // 检查用户是否开启了"减少动画"偏好设置（可访问性）
    const prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches

    // 检查浏览器是否支持 View Transitions API
    type ViewTransitionLike = {
      ready: Promise<void>
      finished: Promise<void>
      updateCallbackDone?: Promise<void>
    }
    const startViewTransition = (
      document as Document & {
        startViewTransition?: (callback: () => void) => ViewTransitionLike
      }
    ).startViewTransition?.bind(document)

    if (prefersReducedMotion || !startViewTransition) {
      // 降级处理：直接切换，无动画
      applyThemeToggle()
      return
    }

    isTransitioning = true

    // 使用 View Transitions API
    const transition = startViewTransition(() => {
      applyThemeToggle()
    })

    await transition.ready

    // 统一使用扩散动画：新主题从点击位置向外扩散覆盖旧主题
    const animation = document.documentElement.animate(
      {
        clipPath: [`circle(0px at ${x}px ${y}px)`, `circle(${endRadius}px at ${x}px ${y}px)`],
      },
      {
        duration: 500,
        easing: 'cubic-bezier(0.4, 0, 0.2, 1)',
        pseudoElement: '::view-transition-new(root)',
      },
    )

    // 等待动画完成
    await animation.finished
    await transition.finished
    isTransitioning = false
  }

  const applyTheme = () => {
    switch (mode.value) {
      case 'light':
        theme.value = 'light'
        break
      case 'dark':
        theme.value = 'dark'
        break
      case 'sunny':
        theme.value = 'sunny'
        break
    }

    document.documentElement.classList.remove('light', 'dark', 'sunny')
    document.documentElement.classList.add(theme.value)
    syncThemeColorMeta()
    localStg.setItem('theme', theme.value)
  }

  const syncThemeColorMeta = () => {
    const rootStyles = getComputedStyle(document.documentElement)
    const canvasColor = rootStyles.getPropertyValue('--color-bg-canvas').trim()
    const nextThemeColor = canvasColor || THEME_COLOR_FALLBACK[theme.value]

    let themeColorMeta = document.querySelector<HTMLMetaElement>(`meta[name="${THEME_COLOR_META_NAME}"]`)
    if (!themeColorMeta) {
      themeColorMeta = document.createElement('meta')
      themeColorMeta.setAttribute('name', THEME_COLOR_META_NAME)
      document.head.appendChild(themeColorMeta)
    }

    themeColorMeta.setAttribute('content', nextThemeColor)
  }

  const init = () => {
    applyTheme()
  }

  return {
    theme,
    mode,
    toggleTheme,
    applyTheme,
    init,
  }
})
