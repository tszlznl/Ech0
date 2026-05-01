// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

export const FILE_CATEGORY = {
  IMAGE: 'image',
  AUDIO: 'audio',
  VIDEO: 'video',
  DOCUMENT: 'document',
  FILE: 'file',
} as const

export type FileCategory = (typeof FILE_CATEGORY)[keyof typeof FILE_CATEGORY]

export const FILE_CATEGORY_VALUES = Object.values(FILE_CATEGORY) as FileCategory[]

export const FILE_STORAGE_TYPE = {
  LOCAL: 'local',
  OBJECT: 'object',
  EXTERNAL: 'external',
} as const

export type FileStorageType = (typeof FILE_STORAGE_TYPE)[keyof typeof FILE_STORAGE_TYPE]

export const FILE_STORAGE_TYPE_VALUES = Object.values(FILE_STORAGE_TYPE) as FileStorageType[]
