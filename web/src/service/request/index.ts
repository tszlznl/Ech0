// 封装ofetch

import { ofetch } from 'ofetch'
import { getAuthToken, getInitReadyStatus } from './shared'
import { theToast } from '@/utils/toast'
import { i18n } from '@/locales'

interface RequestOptions {
  dirrectUrl?: string
  dirrectUrlAndData?: string
  url: string
  method: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'
  timeout?: number
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  data?: any
}

const ofetchInstance = ofetch.create({
  baseURL: import.meta.env.VITE_SERVICE_BASE_URL,
  timeout: 20000,
  ignoreResponseError: true, // 忽略响应错误，让响应拦截器处理

  // 请求拦截器
  onRequest({ options }) {
    const token = getAuthToken()
    const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC'

    const isDirectUrl = options.headers.get('X-Direct-URL')
    if (token && token.length > 0 && !isDirectUrl) {
      options.headers.append('Authorization', token)
    }
    if (!isDirectUrl) {
      options.headers.set('X-Timezone', timezone)
      options.headers.set('X-Locale', i18n.global.locale.value)
    }

    // 清空请求头
    options.headers.delete('X-Direct-URL')
  },
  // 响应拦截器
  onResponseError: async ({ response }) => {
    let data
    try {
      data = await response.json()
    } catch {
      data = { code: 0, msg: String(i18n.global.t('common.requestFailed')), data: null }
    }

    response._data = data

    // 不再 throw，让后续 then() 也能拿到
    return data
  },
})

export const request = async <T>(requestOptions: RequestOptions): Promise<App.Api.Response<T>> => {
  // 检查系统是否已经准备好
  const isSystemReady = getInitReadyStatus()

  // 检查是否使用正向代理
  if (import.meta.env.VITE_PROXY === 'YES') {
    const proxyUrl = import.meta.env.VITE_PROXY_URL
    if (!proxyUrl) {
      throw new Error('Proxy URL is not defined')
    }
    requestOptions.url = `${proxyUrl}${requestOptions.url}`
  }

  return ofetchInstance<App.Api.Response<T>>(requestOptions.url, {
    method: requestOptions.method,
    body: requestOptions.data,
    timeout: requestOptions.timeout,
  }).then((res) => {
    if (res.code !== 1) {
      if (isSystemReady) {
        const translated =
          res.message_key && i18n.global.te(res.message_key)
            ? i18n.global.t(res.message_key, (res.message_params || {}) as Record<string, unknown>)
            : res.msg
        theToast.error(
          translated ? String(translated) : String(i18n.global.t('common.requestFailed')),
        )
      }
    }

    return res
  })
}

// 直接请求
export const requestWithDirectUrl = async <T>(
  requestOptions: RequestOptions,
): Promise<App.Api.Response<T>> => {
  // 检查系统是否已经准备好
  const isSystemReady = getInitReadyStatus()

  return ofetchInstance<App.Api.Response<T>>(
    requestOptions.dirrectUrl ? requestOptions.dirrectUrl : '',
    {
      method: requestOptions.method,
      body: requestOptions.data,
      timeout: requestOptions.timeout,
    },
  ).then((res) => {
    if (res.code !== 1) {
      if (isSystemReady) {
        const translated =
          res.message_key && i18n.global.te(res.message_key)
            ? i18n.global.t(res.message_key, (res.message_params || {}) as Record<string, unknown>)
            : res.msg
        theToast.error(
          translated ? String(translated) : String(i18n.global.t('common.requestFailed')),
        )
      }
    }

    return res
  })
}

// 直接请求 && 直接传递数据
export const requestWithDirectUrlAndData = async <T>(
  requestOptions: RequestOptions,
): Promise<T> => {
  return ofetchInstance<T>(
    requestOptions.dirrectUrlAndData ? requestOptions.dirrectUrlAndData : '',
    {
      method: requestOptions.method,
      body: requestOptions.data,
      timeout: requestOptions.timeout,
      headers: {
        'X-Direct-URL': requestOptions.dirrectUrlAndData ? requestOptions.dirrectUrlAndData : '',
      },
    },
  ).then((res) => {
    return res
  })
}

export const downloadFile = async (requestOptions: RequestOptions): Promise<Blob> => {
  // 检查是否使用正向代理
  if (import.meta.env.VITE_PROXY === 'YES') {
    const proxyUrl = import.meta.env.VITE_PROXY_URL
    if (!proxyUrl) {
      throw new Error('Proxy URL is not defined')
    }
    requestOptions.url = `${proxyUrl}${requestOptions.url}`
  }

  return ofetchInstance<Blob>(requestOptions.url, {
    method: requestOptions.method,
    body: requestOptions.data,
    timeout: requestOptions.timeout,
  }).then((res) => {
    if (res instanceof Blob) {
      return res
    }
    const msg = String(i18n.global.t('common.requestFailed'))
    theToast.error(msg)
    throw new Error(msg)
  })
}
