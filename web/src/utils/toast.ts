// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// 处理全局通知工具函数 (基于 vue-sonner)
import { toast } from 'vue-sonner'

// 定义自定义通知选项接口
interface customToastOptions {
  /** 相同 id 会替换已有 toast，避免同一错误连弹多次 */
  id?: string | number
  duration?: number
  description?: string
  action?: {
    label?: string
    onClick?: () => void
  }
  classes?: {
    actionButton?: string
  }
}

// 默认通知选项
const defaultToastOptions: customToastOptions = {
  duration: 1200, // 默认持续时间为1200毫秒
  description: '', // 默认描述为空
  classes: {
    actionButton: 'bg-blue-500 text-white hover:bg-blue-600 focus:ring-blue-500',
  },
}

function show(
  type: 'success' | 'error' | 'info' | 'warning',
  content: string,
  options?: customToastOptions,
) {
  toast[type](content, {
    duration: options?.duration ?? defaultToastOptions.duration,
    description: options?.description ?? defaultToastOptions.description,
    ...(options?.id !== undefined ? { id: options.id } : {}),
    ...(options?.action?.label
      ? {
          action: {
            label: options.action.label,
            onClick: options.action.onClick ?? (() => toast.dismiss()),
          },
        }
      : {}),
  })
}

// 简化后的 promise toast
function showPromise<T>(
  promise: Promise<T>,
  messages: {
    loading: string
    success: string | ((data: T) => string)
    error: string | ((error: string) => string)
  },
  options?: customToastOptions,
): Promise<T> {
  toast.promise(promise, {
    loading: messages.loading,
    success: messages.success,
    error: messages.error,
    duration: options?.duration ?? defaultToastOptions.duration,
  })
  return promise
}

export const theToast = {
  success: (content: string, options?: customToastOptions) => show('success', content, options),
  error: (content: string, options?: customToastOptions) => show('error', content, options),
  info: (content: string, options?: customToastOptions) => show('info', content, options),
  warning: (content: string, options?: customToastOptions) => show('warning', content, options),
  promise: showPromise,
}
