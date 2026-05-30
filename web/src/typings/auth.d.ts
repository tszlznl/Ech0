// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// 认证 / 登录 / Passkey 相关类型（通过命名空间合并扩展 App.Api）。
declare namespace App {
  namespace Api {
    namespace Auth {
      type LoginParams = {
        username: string
        password: string
      }

      type LoginResponse = {
        access_token: string
        expires_in: number
      }

      type TokenPairResponse = {
        access_token: string
        expires_in: number
      }

      type SignupParams = {
        username: string
        password: string
        email?: string
        locale?: string
      }

      // Passkey / WebAuthn
      type PasskeyRegisterBeginResp = {
        nonce: string
        publicKey: unknown
      }

      type PasskeyLoginBeginResp = {
        nonce: string
        publicKey: unknown
      }

      type PasskeyDevice = {
        id: string
        device_name: string
        aaguid: string
        last_used_at: number
        created_at: number
      }
    }
  }
}
