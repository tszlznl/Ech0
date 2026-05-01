// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { afterEach, vi } from 'vitest'

const createStorageMock = () => {
  const storage = new Map<string, string>()
  return {
    getItem: (key: string) => storage.get(key) ?? null,
    setItem: (key: string, value: string) => {
      storage.set(key, String(value))
    },
    removeItem: (key: string) => {
      storage.delete(key)
    },
    clear: () => {
      storage.clear()
    },
    key: (index: number) => Array.from(storage.keys())[index] ?? null,
    get length() {
      return storage.size
    },
  }
}

vi.stubGlobal('localStorage', createStorageMock())
vi.stubGlobal('sessionStorage', createStorageMock())

afterEach(() => {
  vi.unstubAllGlobals()
  vi.stubGlobal('localStorage', createStorageMock())
  vi.stubGlobal('sessionStorage', createStorageMock())
})
