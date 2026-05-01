// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

/**
 * Hub 仅使用 **light** 语义主题（`web` 的 `:root.light`，见 themes/tokens/semantic.light.scss）。
 * 不读取 localStorage、不响应 dark / sunny，避免与主站主题设置串台。
 */
const THEME_LIGHT = 'light' as const

const THEME_COLOR_FALLBACK_LIGHT = '#f4f1ec'

function syncThemeColorMeta(): void {
  const apply = () => {
    const rootStyles = getComputedStyle(document.documentElement)
    const chromeColor = rootStyles.getPropertyValue('--color-chrome-theme').trim()
    const canvasColor = rootStyles.getPropertyValue('--color-bg-canvas').trim()
    const next = chromeColor || canvasColor || THEME_COLOR_FALLBACK_LIGHT
    let meta = document.querySelector<HTMLMetaElement>('meta[name="theme-color"]')
    if (!meta) {
      meta = document.createElement('meta')
      meta.setAttribute('name', 'theme-color')
      document.head.appendChild(meta)
    }
    meta.setAttribute('content', next)
  }
  requestAnimationFrame(apply)
}

/** 始终为 `<html>` 挂上 `light`，并同步 `theme-color` meta。 */
export function applyWebRootThemeClass(): typeof THEME_LIGHT {
  const root = document.documentElement
  root.classList.remove('light', 'dark', 'sunny')
  root.classList.add(THEME_LIGHT)
  syncThemeColorMeta()
  return THEME_LIGHT
}

if (typeof document !== 'undefined') {
  applyWebRootThemeClass()
}
