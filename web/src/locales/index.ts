// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { createI18n } from 'vue-i18n'
import { localStg } from '@/utils/storage'

export const LOCALE_STORAGE_KEY = 'locale'
// DEFAULT_LOCALE 是源语言（项目原文为中文），用作 vue-i18n 翻译缺失时的回退。
// FALLBACK_LOCALE 是「检测到的语言不在支持列表里」时的兜底，更国际化。
export const DEFAULT_LOCALE = 'zh-CN'
export const FALLBACK_LOCALE = 'en-US'
export const SUPPORTED_LOCALES = ['zh-CN', 'en-US', 'de-DE', 'ja-JP'] as const

export type AppLocale = (typeof SUPPORTED_LOCALES)[number]

const loadedLocales = new Set<string>()

// toSupported 把任意输入映射到受支持的 locale；命中不了就返回 null，
// 让上层的优先级阶梯把判断交给下一个候选（关键：不要提前兜底成 FALLBACK，
// 否则一个不被支持但非空的浏览器语言会短路掉「站点默认」这一层）。
const toSupported = (raw?: string | null): AppLocale | null => {
  const value = String(raw || '').trim()
  if (!value) return null

  const exact = SUPPORTED_LOCALES.find((locale) => locale.toLowerCase() === value.toLowerCase())
  if (exact) return exact

  const langPrefix = value.slice(0, 2).toLowerCase()
  if (langPrefix === 'en') return 'en-US'
  if (langPrefix === 'zh') return 'zh-CN'
  if (langPrefix === 'de') return 'de-DE'
  if (langPrefix === 'ja') return 'ja-JP'

  return null // 不支持 → 交给下一个候选
}

const normalizeLocale = (raw?: string | null): AppLocale => toSupported(raw) ?? FALLBACK_LOCALE

// 模块加载时的初值：本设备记忆 > 浏览器语言 > 回退（站点默认此刻还拿不到，
// 由随后 await 的 setupI18n 用完整阶梯精修，挂载前就会落定）。
const initialLocale =
  toSupported(localStg.getItem<string>(LOCALE_STORAGE_KEY)) ||
  toSupported(navigator.languages?.[0] || navigator.language) ||
  FALLBACK_LOCALE

export const i18n = createI18n({
  legacy: false,
  globalInjection: true,
  locale: initialLocale,
  fallbackLocale: DEFAULT_LOCALE,
  messages: {},
})

async function loadLocaleMessages(locale: AppLocale) {
  if (loadedLocales.has(locale)) return
  const messages = await import(`./messages/${locale}.json`)
  i18n.global.setLocaleMessage(locale, messages.default)
  loadedLocales.add(locale)
}

export async function setI18nLocale(locale: string): Promise<AppLocale> {
  const normalized = normalizeLocale(locale)
  await loadLocaleMessages(normalized)
  i18n.global.locale.value = normalized
  document.documentElement.setAttribute('lang', normalized)
  localStg.setItem(LOCALE_STORAGE_KEY, normalized)
  return normalized
}

export async function setupI18n(defaultLocale?: string) {
  const fromStorage = localStg.getItem<string>(LOCALE_STORAGE_KEY)
  const fromNavigator = navigator.languages?.[0] || navigator.language
  // 阶梯：本设备显式选择 > 浏览器语言 > 站点默认 > 回退。
  // 登录用户的 user.locale 由登录流程（stores/user.ts）先行写入 localStorage，
  // 故在此作为最高层经由 fromStorage 命中，无需在这里单独处理。
  const locale =
    toSupported(fromStorage) ||
    toSupported(fromNavigator) ||
    toSupported(defaultLocale) ||
    FALLBACK_LOCALE
  await setI18nLocale(locale)
  return i18n
}
