// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { FILE_STORAGE_TYPE } from '@/constants/file'
// 直接引 selectors 而非 @/lib/file barrel：让本模块（及经由它的 TheMediaPlayer）
// 不牵连整个 file 子系统，从而可被隔离 bundle 的 hub 项目安全复用。
import { filterFiles, type FileSelectorOptions } from '@/lib/file/selectors/file-selectors'

type EchoLike = {
  id?: string
  layout?: string | null
  echo_files?: Array<{
    echo_id?: string
    file_id?: string
    file?: {
      id?: string
      key?: string
      url?: string
      storage_type?: string
      category?: string
      content_type?: string
      size?: number
      width?: number
      height?: number
    }
  }> | null
}

export function getEchoFiles(echo?: EchoLike | null): App.Api.Ech0.FileObject[] {
  if (!echo) return []

  return (echo.echo_files || []).map((item) => {
    const file = item.file
    const storageType = normalizeStorageType(file?.storage_type)
    return {
      id: String(file?.id || item.file_id || ''),
      echo_id: String(echo.id || item.echo_id || ''),
      url: String(file?.url || ''),
      storage_type: storageType,
      category: (file?.category as App.Api.File.Category | undefined) || undefined,
      content_type: file?.content_type,
      key: String(file?.key || ''),
      size: file?.size,
      width: file?.width,
      height: file?.height,
    }
  })
}

export function getEchoFilesBy(
  echo?: EchoLike | null,
  options: FileSelectorOptions<App.Api.Ech0.FileObject> = {},
): App.Api.Ech0.FileObject[] {
  return filterFiles(getEchoFiles(echo), options)
}

// backward-compatible alias
export const getEchoImages = (echo?: EchoLike | null) =>
  getEchoFilesBy(echo, { categories: ['image'] })

// 对应 ImageLayout.GRID / HORIZONTAL / STACK。
// 这里刻意用字面量而非引入 ImageLayout 枚举，保持本模块可被隔离 bundle 的 hub 项目安全复用。
const CONTENT_LEADING_LAYOUTS = ['grid', 'horizontal', 'stack']

/**
 * 卡片里正文是否应排在媒体上方（媒体在下）：仅 GRID / HORIZONTAL / STACK 这几种图片布局如此。
 * 其余情况（瀑布流 / 单图轮播、音频、视频）都是媒体在上、正文在下。
 */
export function isContentLeadingEcho(echo?: EchoLike | null): boolean {
  const layout = echo?.layout
  return !!layout && CONTENT_LEADING_LAYOUTS.includes(layout)
}

function normalizeStorageType(raw: unknown): App.Api.File.StorageType {
  const value = String(raw || '').toLowerCase()
  if (value === FILE_STORAGE_TYPE.OBJECT) return FILE_STORAGE_TYPE.OBJECT
  if (value === FILE_STORAGE_TYPE.EXTERNAL) return FILE_STORAGE_TYPE.EXTERNAL
  return FILE_STORAGE_TYPE.LOCAL
}

// 估算 markdown 内容的"字数"。中文按字符计、英文按空白分词。
// 目标只是给读者一个量级，不追求精确。
export function countWords(content: string | null | undefined): number {
  if (!content) return 0
  const stripped = content
    .replace(/```[\s\S]*?```/g, ' ')
    .replace(/`[^`]*`/g, ' ')
    .replace(/!\[[^\]]*\]\([^)]*\)/g, ' ')
    .replace(/\[([^\]]*)\]\([^)]*\)/g, '$1')
    .replace(/[#>*_~`[\]()!\-]+/g, ' ')
    .trim()
  if (!stripped) return 0
  const cjk = stripped.match(/[一-鿿぀-ヿ가-힯]/g)?.length ?? 0
  const ascii = stripped
    .replace(/[一-鿿぀-ヿ가-힯]/g, ' ')
    .split(/\s+/)
    .filter(Boolean).length
  return cjk + ascii
}
