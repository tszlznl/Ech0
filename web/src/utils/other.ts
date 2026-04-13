import { MusicProvider } from '@/enums/enums'
import { i18n, DEFAULT_LOCALE } from '@/locales'

const ABSOLUTE_URL_REGEX = /^https?:\/\//i
const joinBaseAndPath = (baseUrl: string, path: string) =>
  `${baseUrl.replace(/\/+$/, '')}/${path.replace(/^\/+/, '')}`
const defaultServiceBaseUrl = String(import.meta.env.VITE_SERVICE_BASE_URL || '').trim()

const normalizeMediaPath = (path: string) => {
  if (path.startsWith('/api/') || path.startsWith('api/')) return path
  if (path.startsWith('/files/') || path.startsWith('files/'))
    return `/api/${path.replace(/^\/+/, '')}`
  return path
}

const resolveFileUrlByPath = (rawUrl?: string, baseUrl?: string) => {
  const candidate = String(rawUrl ?? '').trim()
  if (!candidate || ABSOLUTE_URL_REGEX.test(candidate)) return candidate
  const base = String(baseUrl ?? defaultServiceBaseUrl).trim()
  const path = normalizeMediaPath(candidate)
  return base ? joinBaseAndPath(base, path) : path
}

const resolveFileUrl = (
  file: Pick<App.Api.Ech0.FileObject | App.Api.Ech0.FileToAdd, 'url'> & { image_url?: string },
  baseUrl?: string,
) => resolveFileUrlByPath(file.url || file.image_url, baseUrl)

// 获取图片链接
export const getFileUrl = (file: App.Api.Ech0.FileObject) => resolveFileUrl(file)

// 获取待添加图片链接
export const getFileToAddUrl = (file: App.Api.Ech0.FileToAdd) => resolveFileUrl(file)

// backward-compatible aliases
export const getImageUrl = (image: App.Api.Ech0.FileObject) => getFileUrl(image)
export const getImageToAddUrl = (image: App.Api.Ech0.FileToAdd) => getFileToAddUrl(image)

export const formatDate = (dateInput: string | number) => {
  // 同一本地日历日：刚刚 / N 分钟前 / N 小时前（不按 24h 粗算「天」）
  // 相差 1～2 个本地日历日：N 天前（与「昨天」语义一致，避免昨夜帖子仍显示「N 小时前」）
  // 更早：Intl 按 locale 显示完整日期

  // 处理 Unix 时间戳（秒或毫秒）和日期字符串
  let date: Date
  if (typeof dateInput === 'number') {
    // 如果是数字，判断是秒级还是毫秒级时间戳
    // 秒级时间戳通常小于 10^12，毫秒级时间戳通常大于 10^12
    date = new Date(dateInput < 1e12 ? dateInput * 1000 : dateInput)
  } else {
    date = new Date(dateInput)
  }

  const now = new Date()
  const diff = now.getTime() - date.getTime()

  const locale = i18n.global.locale.value || DEFAULT_LOCALE
  const t = (key: string, params?: Record<string, unknown>) =>
    String(i18n.global.t(key, params || {}))

  const localMidnight = (d: Date) => new Date(d.getFullYear(), d.getMonth(), d.getDate())
  const calendarDaysFromDateToNow = Math.round(
    (localMidnight(now).getTime() - localMidnight(date).getTime()) / (1000 * 60 * 60 * 24),
  )

  const longFormatter = () =>
    new Intl.DateTimeFormat(locale, {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      weekday: 'short',
    }).format(date)

  if (diff < 0) {
    return longFormatter()
  }

  const diffInSeconds = Math.floor(diff / 1000)
  const diffInMinutes = Math.floor(diff / (1000 * 60))
  const diffInHours = Math.floor(diff / (1000 * 60 * 60))

  if (calendarDaysFromDateToNow === 0) {
    if (diffInSeconds < 60) {
      return t('dateTime.justNow')
    }
    if (diffInMinutes < 60) {
      return t('dateTime.minutesAgo', { count: diffInMinutes })
    }
    return t('dateTime.hoursAgo', { count: diffInHours })
  }

  if (calendarDaysFromDateToNow === 1 || calendarDaysFromDateToNow === 2) {
    return t('dateTime.daysAgo', { count: calendarDaysFromDateToNow })
  }

  return longFormatter()
}

// 解析音乐链接（网易云、QQ音乐、Apple Music）
export const parseMusicURL = (url: string) => {
  url = url.trim()

  /* =======================
   * 网易云音乐
   * ======================= */
  if (/^https:\/\/([a-z0-9-]+\.)*music\.163\.com/i.test(url)) {
    // id：永远从 query 拿
    const idMatch = url.match(/[?&]id=(\d+)/)
    if (!idMatch) return null

    let type: 'song' | 'playlist' | 'album' | undefined

    if (/(\/|#\/|\/m\/)song/.test(url)) {
      type = 'song'
    } else if (/(\/|#\/|\/m\/)playlist/.test(url)) {
      type = 'playlist'
    }
    // else if (/(\/|#\/|\/m\/)album/.test(url)) {
    //   type = 'album'
    // }

    if (!type) return null

    return {
      server: MusicProvider.NETEASE,
      type,
      id: idMatch[1],
    }
  }

  /* =======================
   * QQ 音乐
   * ======================= */
  if (/^https:\/\/([a-z0-9-]+\.)*qq\.com/i.test(url)) {
    // 新版 songDetail（字母数字混合）
    const newSongMatch = url.match(/songDetail\/([a-zA-Z0-9]+)/)
    if (newSongMatch) {
      return {
        server: MusicProvider.QQ,
        type: 'song',
        id: newSongMatch[1],
      }
    }

    // 旧版 songid（纯数字）
    const oldSongMatch = url.match(/[?&]songid=(\d+)/)
    if (oldSongMatch) {
      return {
        server: MusicProvider.QQ,
        type: 'song',
        id: oldSongMatch[1],
      }
    }

    // 匹配 /playlist/ 后面的数字ID
    const playlistMatch = url.match(/\/playlist\/(\d+)/i)
    if (playlistMatch) {
      return {
        server: MusicProvider.QQ,
        type: 'playlist',
        id: playlistMatch[1],
      }
    }
    return null
  }

  /* =======================
   * Apple Music
   * ======================= */
  if (/^https:\/\/music\.apple\.com/i.test(url)) {
    const appleMatch = url.match(/\/(song|album)\/[^/]+\/(\d+)/)
    if (!appleMatch) return null

    return {
      server: MusicProvider.APPLE,
      type: appleMatch[1],
      id: appleMatch[2],
    }
  }
  return null
}

/**
 * 从一段文本中提取并返回最短、规范的音乐链接
 */
export const extractAndCleanMusicURL = (input: string): string | null => {
  const text = input.trim()

  // 粗暴提取第一个 URL（足够鲁棒）
  const urlMatch = text.match(/https?:\/\/[^\s]+/i)
  if (!urlMatch) return null

  const rawUrl = urlMatch[0]

  // 复用统一解析函数
  const parsed = parseMusicURL(rawUrl)
  if (!parsed) return null

  // 规范化重组
  switch (parsed.server) {
    case MusicProvider.NETEASE: {
      // 统一为 PC 可打开的最短链接
      return `https://music.163.com/#/${parsed.type}?id=${parsed.id}`
    }

    case MusicProvider.QQ: {
      // 根据 type 分别生成歌曲或歌单的规范化链接
      if (parsed.type === 'song') {
        return `https://y.qq.com/n/ryqq_v2/songDetail/${parsed.id}`
      }

      if (parsed.type === 'playlist') {
        // 使用统一的歌单路径格式
        return `https://y.qq.com/n/ryqq_v2/playlist/${parsed.id}`
      }

      // 如果是其他未知类型，则返回 null
      return null
    }

    case MusicProvider.APPLE: {
      // 去掉后面的参数
      const cleanUrl = rawUrl.split('?')[0]
      return cleanUrl ?? rawUrl
    }

    default:
      return null
  }
}

// 获取 HubEcho 的图片
export const getHubImageUrl = (image: App.Api.Ech0.FileObject, baseurl: string) => {
  return resolveFileUrl(image, baseurl)
}

/**
 * Base64URL to Uint8Array
 * 用于解析服务端返回的 WebAuthn publicKey
 */
export function base64urlToUint8Array(input: string): Uint8Array {
  const base64 = input.replace(/-/g, '+').replace(/_/g, '/')
  const pad = base64.length % 4 === 0 ? '' : '='.repeat(4 - (base64.length % 4))
  const binary = atob(base64 + pad)
  const buffer = new ArrayBuffer(binary.length)
  const bytes = new Uint8Array(buffer)
  for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i)
  return bytes
}

/**
 * Uint8Array to Base64URL
 * 用于生成客户端返回的 WebAuthn publicKey
 */
export function uint8ArrayToBase64url(bytes: ArrayBuffer | Uint8Array): string {
  const u8 = bytes instanceof Uint8Array ? bytes : new Uint8Array(bytes)
  let binary = ''
  for (let i = 0; i < u8.length; i++) binary += String.fromCharCode(u8[i]!)
  const base64 = btoa(binary)
  return base64.replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/g, '')
}

export function isSafari(): boolean {
  const ua = navigator.userAgent
  // Exclude Chrome, Chromium-based Edge, and iOS Chrome/Firefox
  return (
    ua.includes('Safari') &&
    !ua.includes('Chrome') &&
    !ua.includes('CriOS') &&
    !ua.includes('FxiOS')
  )
}
