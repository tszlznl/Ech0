// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

declare module 'vue-virtual-scroller' {
  import type { DefineComponent } from 'vue'

  export const DynamicScroller: DefineComponent<
    Record<string, unknown>,
    Record<string, unknown>,
    unknown
  >
  export const DynamicScrollerItem: DefineComponent<
    Record<string, unknown>,
    Record<string, unknown>,
    unknown
  >
}
