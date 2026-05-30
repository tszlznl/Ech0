// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// API 通用响应包装类型（App.Api 根级公共类型）。
declare namespace App {
  /**
   * Namespace Api
   */
  namespace Api {
    type Response<T> = {
      code: number
      msg: string
      error_code?: string
      message_key?: string
      message_params?: Record<string, unknown>
      data: T
    }
  }
}
