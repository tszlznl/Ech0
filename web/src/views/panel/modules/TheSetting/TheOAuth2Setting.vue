<template>
  <div>
    <PanelCard class="mb-3">
      <!-- OAuth2 设置 -->
      <div class="w-full">
        <div class="flex flex-row items-center justify-between mb-3">
          <h1 class="text-[var(--color-text-primary)] font-bold text-lg">OAuth2设置</h1>
          <div class="flex flex-row items-center justify-end">
            <BaseEditCapsule
              :editing="oauth2EditMode"
              apply-title="应用"
              cancel-title="取消"
              edit-title="编辑"
              @apply="handleUpdateOAuth2Setting"
              @toggle="oauth2EditMode = !oauth2EditMode"
            />
          </div>
        </div>

        <div
          v-if="OAuth2Setting.enable"
          class="mb-3 border border-dashed border-[var(--color-border-strong)] rounded-md p-3"
        >
          <h2 class="text-[var(--color-text-primary)] font-semibold mb-2">配置健康检查</h2>
          <div class="flex flex-col sm:flex-row sm:flex-wrap gap-2 text-sm">
            <div class="flex items-center gap-2">
              <span class="text-[var(--color-text-secondary)]">OAuth就绪:</span>
              <span
                class="px-2 py-0.5 rounded-md"
                :class="
                  oauthRuntimeStatus?.oauth_ready
                    ? 'bg-green-500/15 text-green-500'
                    : 'bg-yellow-500/15 text-yellow-500'
                "
              >
                {{ oauthRuntimeStatus?.oauth_ready ? '已就绪' : '未就绪' }}
              </span>
            </div>
            <div class="flex items-center gap-2">
              <span class="text-[var(--color-text-secondary)]">Passkey就绪:</span>
              <span
                class="px-2 py-0.5 rounded-md"
                :class="
                  oauthRuntimeStatus?.passkey_ready
                    ? 'bg-green-500/15 text-green-500'
                    : 'bg-yellow-500/15 text-yellow-500'
                "
              >
                {{ oauthRuntimeStatus?.passkey_ready ? '已就绪' : '未就绪' }}
              </span>
            </div>
          </div>
          <p
            v-if="missingBoundaryItems.length > 0"
            class="mt-2 text-xs text-[var(--color-text-muted)] break-all"
          >
            缺失项: {{ missingBoundaryItems.join('、') }}
          </p>
          <div class="mt-2">
            <BaseButton
              class="rounded-md h-8 text-xs"
              @click="handleAutoFillBoundary"
              :disabled="missingBoundaryItems.length === 0"
            >
              一键填充推荐配置
            </BaseButton>
          </div>
        </div>

        <!-- 开启OAuth2 -->
        <div
          class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] h-10"
        >
          <h2 class="font-semibold w-30 shrink-0">启用OAuth2:</h2>
          <BaseSwitch v-model="OAuth2Setting.enable" :disabled="!oauth2EditMode" />
        </div>

        <!-- OAuth2 Provider -->
        <div
          class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 h-10"
        >
          <h2 class="font-semibold w-30 shrink-0">OAuth2 模板:</h2>
          <BaseSelect
            v-model="OAuth2Setting.provider"
            :options="OAuth2ProviderOptions"
            :disabled="!oauth2EditMode"
            class="w-34 h-8"
          />
        </div>

        <!-- Client ID -->
        <div
          class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 h-10"
        >
          <h2 class="font-semibold w-30 shrink-0">Client ID:</h2>
          <span
            v-if="!oauth2EditMode"
            class="truncate max-w-40 inline-block align-middle"
            :title="OAuth2Setting.client_id"
            style="vertical-align: middle"
          >
            {{ OAuth2Setting.client_id.length === 0 ? '暂无' : OAuth2Setting.client_id }}
          </span>
          <BaseInput
            v-else
            v-model="OAuth2Setting.client_id"
            type="text"
            placeholder="请输入Client ID"
            class="w-full py-1!"
          />
        </div>

        <!-- Client Secret -->
        <div
          class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 h-10"
        >
          <h2 class="font-semibold w-30 shrink-0">Client Secret:</h2>
          <span
            v-if="!oauth2EditMode"
            class="truncate max-w-40 inline-block align-middle"
            :title="OAuth2Setting.client_secret"
            style="vertical-align: middle"
          >
            {{ OAuth2Setting.client_secret.length === 0 ? '暂无' : OAuth2Setting.client_secret }}
          </span>
          <BaseInput
            v-else
            v-model="OAuth2Setting.client_secret"
            type="text"
            placeholder="请输入Client Secret"
            class="w-full py-1!"
          />
        </div>

        <!-- Callback URL -->
        <div
          class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 h-10"
        >
          <h2 class="font-semibold w-30 shrink-0">Callback URL:</h2>
          <span
            v-if="!oauth2EditMode"
            class="truncate max-w-40 inline-block align-middle"
            :title="OAuth2Setting.redirect_uri"
            style="vertical-align: middle"
          >
            {{ redirect_uri.length === 0 ? '暂无' : redirect_uri }}
          </span>
          <BaseInput
            v-else
            v-model="redirect_uri"
            type="text"
            placeholder="请输入回调地址"
            class="w-full py-1!"
          />
        </div>

        <!-- Auth URL -->
        <div
          class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 h-10"
        >
          <h2 class="font-semibold w-30 shrink-0">Auth URL:</h2>
          <span
            v-if="!oauth2EditMode"
            class="truncate max-w-40 inline-block align-middle"
            :title="OAuth2Setting.auth_url"
            style="vertical-align: middle"
          >
            {{ OAuth2Setting.auth_url.length === 0 ? '暂无' : OAuth2Setting.auth_url }}
          </span>
          <BaseInput
            v-else
            v-model="OAuth2Setting.auth_url"
            type="text"
            placeholder="请输入授权地址"
            class="w-full py-1!"
          />
        </div>

        <!-- Token URL -->
        <div
          class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 h-10"
        >
          <h2 class="font-semibold w-30 shrink-0">Token URL:</h2>
          <span
            v-if="!oauth2EditMode"
            class="truncate max-w-40 inline-block align-middle"
            :title="OAuth2Setting.token_url"
            style="vertical-align: middle"
          >
            {{ OAuth2Setting.token_url.length === 0 ? '暂无' : OAuth2Setting.token_url }}
          </span>
          <BaseInput
            v-else
            v-model="OAuth2Setting.token_url"
            type="text"
            placeholder="请输入Token地址"
            class="w-full py-1!"
          />
        </div>

        <!-- User Info URL -->
        <div
          class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 h-10"
        >
          <h2 class="font-semibold w-30 shrink-0">UserInfo URL:</h2>
          <span
            v-if="!oauth2EditMode"
            class="truncate max-w-40 inline-block align-middle"
            :title="OAuth2Setting.user_info_url"
            style="vertical-align: middle"
          >
            {{ OAuth2Setting.user_info_url.length === 0 ? '暂无' : OAuth2Setting.user_info_url }}
          </span>
          <BaseInput
            v-else
            v-model="OAuth2Setting.user_info_url"
            type="text"
            placeholder="请输入用户信息地址"
            class="w-full py-1!"
          />
        </div>

        <!-- Scopes -->
        <div
          class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 h-10"
        >
          <h2 class="font-semibold w-30 shrink-0">Scopes:</h2>
          <span
            v-if="!oauth2EditMode"
            class="truncate max-w-40 inline-block align-middle"
            :title="OAuth2Setting.scopes.join(', ')"
            style="vertical-align: middle"
          >
            {{ OAuth2Setting.scopes.length === 0 ? '暂无' : OAuth2Setting.scopes.join(', ') }}
          </span>
          <BaseInput
            v-else
            v-model="scopeString"
            type="text"
            placeholder="请输入Scopes，多个用逗号分隔"
            class="w-full py-1!"
            @blur="OAuth2Setting.scopes = scopeString.split(',').map((s) => s.trim())"
          />
        </div>

        <!-- Is OIDC -->
        <div
          v-if="OAuth2Setting.enable"
          class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] h-10"
        >
          <h2 class="font-semibold w-30 shrink-0">启用OIDC:</h2>
          <BaseSwitch v-model="OAuth2Setting.is_oidc" :disabled="!oauth2EditMode" />
        </div>

        <!-- Issuer -->
        <div
          v-if="OAuth2Setting.is_oidc"
          class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 h-10"
        >
          <h2 class="font-semibold w-30 shrink-0">Issuer:</h2>
          <span
            v-if="!oauth2EditMode"
            class="truncate max-w-40 inline-block align-middle"
            :title="OAuth2Setting.issuer"
            style="vertical-align: middle"
          >
            {{ OAuth2Setting.issuer.length === 0 ? '暂无' : OAuth2Setting.issuer }}
          </span>
          <BaseInput
            v-else
            v-model="OAuth2Setting.issuer"
            type="text"
            placeholder="请输入Issuer"
            class="w-full py-1!"
          />
        </div>

        <!-- JWKS URL -->
        <div
          v-if="OAuth2Setting.is_oidc"
          class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 h-10"
        >
          <h2 class="font-semibold w-30 shrink-0">JWKS URL:</h2>
          <span
            v-if="!oauth2EditMode"
            class="truncate max-w-40 inline-block align-middle"
            :title="OAuth2Setting.jwks_url"
            style="vertical-align: middle"
          >
            {{ OAuth2Setting.jwks_url.length === 0 ? '暂无' : OAuth2Setting.jwks_url }}
          </span>
          <BaseInput
            v-else
            v-model="OAuth2Setting.jwks_url"
            type="text"
            placeholder="请输入JWKS URL"
            class="w-full py-1!"
          />
        </div>

        <!-- 认证安全边界（Panel 主配置） -->
        <div class="mt-3 border border-dashed border-[var(--color-border-strong)] rounded-md p-3">
          <h3 class="text-[var(--color-text-primary)] font-semibold mb-2">认证安全边界</h3>
          <div class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 h-10">
            <h2 class="font-semibold w-40 shrink-0">Redirect Allowlist:</h2>
            <span v-if="!oauth2EditMode" class="truncate max-w-80 inline-block align-middle">
              {{
                OAuth2Setting.auth_redirect_allowed_return_urls.length === 0
                  ? '暂无'
                  : OAuth2Setting.auth_redirect_allowed_return_urls.join(', ')
              }}
            </span>
            <BaseInput
              v-else
              v-model="redirectAllowlistString"
              type="text"
              placeholder="多个URL用逗号分隔"
              class="w-full py-1!"
            />
          </div>
          <div class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 h-10">
            <h2 class="font-semibold w-40 shrink-0">WebAuthn RP ID:</h2>
            <span v-if="!oauth2EditMode" class="truncate max-w-80 inline-block align-middle">
              {{ OAuth2Setting.webauthn_rp_id || '暂无' }}
            </span>
            <BaseInput
              v-else
              v-model="OAuth2Setting.webauthn_rp_id"
              type="text"
              placeholder="例如：example.com"
              class="w-full py-1!"
            />
          </div>
          <div class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 h-10">
            <h2 class="font-semibold w-40 shrink-0">WebAuthn Origins:</h2>
            <span v-if="!oauth2EditMode" class="truncate max-w-80 inline-block align-middle">
              {{
                OAuth2Setting.webauthn_allowed_origins.length === 0
                  ? '暂无'
                  : OAuth2Setting.webauthn_allowed_origins.join(', ')
              }}
            </span>
            <BaseInput
              v-else
              v-model="webauthnOriginsString"
              type="text"
              placeholder="多个Origin用逗号分隔"
              class="w-full py-1!"
            />
          </div>
          <div class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 h-10">
            <h2 class="font-semibold w-40 shrink-0">CORS Origins:</h2>
            <span v-if="!oauth2EditMode" class="truncate max-w-80 inline-block align-middle">
              {{
                OAuth2Setting.cors_allowed_origins.length === 0
                  ? '暂无'
                  : OAuth2Setting.cors_allowed_origins.join(', ')
              }}
            </span>
            <BaseInput
              v-else
              v-model="corsOriginsString"
              type="text"
              placeholder="多个Origin用逗号分隔"
              class="w-full py-1!"
            />
          </div>
        </div>
      </div>
    </PanelCard>

    <PanelCard v-if="OAuth2Setting.enable && OAuth2Setting.provider" class="mb-3">
      <!-- OAuth2 账号绑定 -->
      <div class="w-full border border-dashed border-[var(--color-border-strong)] rounded-md p-3">
        <div>
          <h1 class="text-[var(--color-text-primary)] font-semibold text-lg">账号绑定</h1>
          <p class="text-[var(--color-text-muted)] text-sm mt-1">注意：需先配置OAuth2信息</p>
          <div
            v-if="oauthInfo && isBound"
            class="mt-2 border border-dashed border-[var(--color-border-strong)] rounded-md p-3 flex items-center justify-center"
          >
            <p class="text-[var(--color-text-secondary)] font-bold flex items-center">
              <component
                :is="
                  oauthInfo.provider === OAuth2Provider.GITHUB
                    ? Github
                    : oauthInfo.provider === OAuth2Provider.GOOGLE
                      ? Google
                      : oauthInfo.provider === OAuth2Provider.QQ
                        ? QQ
                        : Custom
                "
                class="w-5 h-5 mr-2"
              />
              <span>{{
                oauthInfo.provider === OAuth2Provider.GITHUB
                  ? 'GitHub'
                  : oauthInfo.provider === OAuth2Provider.GOOGLE
                    ? 'Google'
                    : oauthInfo.provider === OAuth2Provider.QQ
                      ? 'QQ'
                      : `自定义 ${oauthInfo?.auth_type === 'oidc' ? 'OIDC' : 'OAuth2'}  账号`
              }}</span>
              已绑定
            </p>
          </div>
          <BaseButton v-else class="rounded-md mt-3" @click="handleBindOAuth2()">
            <div class="flex items-center justify-between">
              <component
                :is="
                  OAuth2Setting.provider === OAuth2Provider.GITHUB
                    ? Github
                    : OAuth2Setting.provider === OAuth2Provider.GOOGLE
                      ? Google
                      : OAuth2Setting.provider === OAuth2Provider.QQ
                        ? QQ
                        : Custom
                "
                class="w-5 h-5 mr-2"
              />
              <span class="flex-1 text-left">
                {{
                  OAuth2Setting.provider === OAuth2Provider.GITHUB
                    ? '绑定 GitHub 账号'
                    : OAuth2Setting.provider === OAuth2Provider.GOOGLE
                      ? '绑定 Google 账号'
                      : OAuth2Setting.provider === OAuth2Provider.QQ
                        ? '绑定 QQ 账号'
                        : `绑定自定义 ${OAuth2Setting.is_oidc ? 'OIDC' : 'OAuth2'} 账号`
                }}
              </span>
            </div>
          </BaseButton>
        </div>
      </div>
    </PanelCard>
  </div>
</template>

<script setup lang="ts">
import PanelCard from '@/layout/PanelCard.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseSwitch from '@/components/common/BaseSwitch.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseEditCapsule from '@/components/common/BaseEditCapsule.vue'
import { ref, onMounted, watch } from 'vue'
import { useSettingStore } from '@/stores'
import { theToast } from '@/utils/toast'
import { OAuth2Provider } from '@/enums/enums'
import {
  fetchUpdateOAuth2Settings,
  fetchBindOAuth2,
  fetchGetOAuthInfo,
  fetchGetOAuth2Status,
} from '@/service/api'
import Github from '@/components/icons/github.vue'
import Google from '@/components/icons/google.vue'
import QQ from '@/components/icons/qq.vue'
import Custom from '@/components/icons/customoauth.vue'
import { storeToRefs } from 'pinia'

const settingStore = useSettingStore()
const { getOAuth2Setting } = settingStore
const { OAuth2Setting } = storeToRefs(settingStore)

const oauth2EditMode = ref(false)

const OAuth2ProviderOptions = [
  { label: 'GitHub', value: OAuth2Provider.GITHUB },
  { label: 'Google', value: OAuth2Provider.GOOGLE },
  { label: 'QQ', value: OAuth2Provider.QQ },
  { label: 'Custom(支持 OIDC)', value: OAuth2Provider.CUSTOM },
]

const redirect_uri = ref<string>(OAuth2Setting.value.redirect_uri)
if (!redirect_uri.value) {
  redirect_uri.value = `${window.location.origin}/oauth/${OAuth2Setting.value.provider}/callback`
}
const scopeString = ref('read:user')
const redirectAllowlistString = ref('')
const webauthnOriginsString = ref('')
const corsOriginsString = ref('')

const parseList = (input: string) =>
  input
    .split(',')
    .map((s) => s.trim())
    .filter((s) => s.length > 0)

const handleUpdateOAuth2Setting = async () => {
  // 修改Scopes
  OAuth2Setting.value.scopes = scopeString.value.split(',').map((s) => s.trim())
  // 修改回调地址为当前域名加上固定路径
  OAuth2Setting.value.redirect_uri =
    redirect_uri.value || `${window.location.origin}/oauth/${OAuth2Setting.value.provider}/callback`
  OAuth2Setting.value.auth_redirect_allowed_return_urls = parseList(redirectAllowlistString.value)
  OAuth2Setting.value.webauthn_allowed_origins = parseList(webauthnOriginsString.value)
  OAuth2Setting.value.cors_allowed_origins = parseList(corsOriginsString.value)

  if (OAuth2Setting.value.auth_redirect_allowed_return_urls.some((u) => !/^https?:\/\//.test(u))) {
    theToast.error('Redirect Allowlist 需为 http/https URL')
    return
  }
  if (OAuth2Setting.value.webauthn_allowed_origins.some((u) => !/^https?:\/\//.test(u))) {
    theToast.error('WebAuthn Origins 需为 http/https URL')
    return
  }
  if (OAuth2Setting.value.cors_allowed_origins.some((u) => !/^https?:\/\//.test(u))) {
    theToast.error('CORS Origins 需为 http/https URL')
    return
  }

  // 提交更新
  await fetchUpdateOAuth2Settings(OAuth2Setting.value)
    .then((res) => {
      if (res.code === 1) {
        theToast.success(res.msg)
      }
    })
    .finally(async () => {
      oauth2EditMode.value = false
      // 重新获取OAuth2设置
      await getOAuth2Setting()
      await refreshHealthCheck()
      // 重新获取OAuth信息
      if (OAuth2Setting.value.provider) {
        const infoRes = await fetchGetOAuthInfo(OAuth2Setting.value.provider)
        if (infoRes.code === 1) {
          oauthInfo.value = infoRes.data
        }
      }
    })
}

const handleBindOAuth2 = async () => {
  const res = await fetchBindOAuth2(OAuth2Setting.value.provider, `${window.location.origin}/panel`)
  if (res.code !== 1) {
    theToast.error(res.msg)
  } else {
    console.log('Bind URL: ', res.data)
    // 成功，跳转到授权URL
    window.location.href = res.data
  }
}

const oauthInfo = ref<App.Api.Setting.OAuthInfo | null>(null)
const isBound = ref<boolean>(false)
const oauthRuntimeStatus = ref<App.Api.Setting.OAuth2Status | null>(null)
const missingBoundaryItems = ref<string[]>([])

const refreshHealthCheck = async () => {
  const statusRes = await fetchGetOAuth2Status()
  if (statusRes.code === 1) {
    oauthRuntimeStatus.value = statusRes.data
  }
  const missing: string[] = []
  if ((OAuth2Setting.value.auth_redirect_allowed_return_urls || []).length === 0) {
    missing.push('Redirect Allowlist')
  }
  if (!OAuth2Setting.value.webauthn_rp_id) {
    missing.push('WebAuthn RP ID')
  }
  if ((OAuth2Setting.value.webauthn_allowed_origins || []).length === 0) {
    missing.push('WebAuthn Origins')
  }
  if ((OAuth2Setting.value.cors_allowed_origins || []).length === 0) {
    missing.push('CORS Origins')
  }
  missingBoundaryItems.value = missing
}

const handleAutoFillBoundary = () => {
  const currentOrigin = window.location.origin
  const currentHost = window.location.hostname
  if (!OAuth2Setting.value.auth_redirect_allowed_return_urls?.length) {
    OAuth2Setting.value.auth_redirect_allowed_return_urls = [`${currentOrigin}/auth`]
  }
  if (!OAuth2Setting.value.webauthn_rp_id) {
    OAuth2Setting.value.webauthn_rp_id = currentHost
  }
  if (!OAuth2Setting.value.webauthn_allowed_origins?.length) {
    OAuth2Setting.value.webauthn_allowed_origins = [currentOrigin]
  }
  if (!OAuth2Setting.value.cors_allowed_origins?.length) {
    OAuth2Setting.value.cors_allowed_origins = [currentOrigin]
  }

  redirectAllowlistString.value = OAuth2Setting.value.auth_redirect_allowed_return_urls.join(', ')
  webauthnOriginsString.value = OAuth2Setting.value.webauthn_allowed_origins.join(', ')
  corsOriginsString.value = OAuth2Setting.value.cors_allowed_origins.join(', ')
  oauth2EditMode.value = true
  void refreshHealthCheck()
  theToast.success('已填充推荐配置，请点击“应用”保存')
}

// 监听 OAuth2Setting.provider 变化，更新必填设置模板
watch(
  () => OAuth2Setting.value.provider,
  (newProvider) => {
    const template = getProviderTemplate(newProvider)
    Object.assign(OAuth2Setting.value, template)
  },
)

watch(
  () => OAuth2Setting.value,
  (v) => {
    redirectAllowlistString.value = (v.auth_redirect_allowed_return_urls || []).join(', ')
    webauthnOriginsString.value = (v.webauthn_allowed_origins || []).join(', ')
    corsOriginsString.value = (v.cors_allowed_origins || []).join(', ')
  },
  { immediate: true, deep: true },
)

function getProviderTemplate(provider: string) {
  if (provider === String(OAuth2Provider.GITHUB)) {
    scopeString.value = 'read:user'
    redirect_uri.value = `${window.location.origin}/oauth/github/callback`
    return {
      redirect_uri: `${window.location.origin}/oauth/github/callback`,
      auth_url: 'https://github.com/login/oauth/authorize',
      token_url: 'https://github.com/login/oauth/access_token',
      user_info_url: 'https://api.github.com/user',
      scopes: ['read:user'],
    }
  } else if (provider === String(OAuth2Provider.GOOGLE)) {
    scopeString.value = 'openid'
    redirect_uri.value = `${window.location.origin}/oauth/google/callback`
    return {
      redirect_uri: `${window.location.origin}/oauth/google/callback`,
      auth_url: 'https://accounts.google.com/o/oauth2/v2/auth',
      token_url: 'https://oauth2.googleapis.com/token',
      user_info_url: 'https://openidconnect.googleapis.com/v1/userinfo',
      scopes: ['openid'], // 只要OAuth ID
    }
  } else if (provider === String(OAuth2Provider.QQ)) {
    scopeString.value = 'get_user_info'
    redirect_uri.value = `${window.location.origin}/oauth/qq/callback`
    return {
      redirect_uri: `${window.location.origin}/oauth/qq/callback`,
      auth_url: 'https://graph.qq.com/oauth2.0/authorize',
      token_url: 'https://graph.qq.com/oauth2.0/token',
      user_info_url: 'https://graph.qq.com/user/get_user_info',
      scopes: ['get_user_info'],
    }
  } else if (provider === String(OAuth2Provider.CUSTOM)) {
    scopeString.value = 'openid'
    redirect_uri.value = `${window.location.origin}/oauth/custom/callback`
    return {
      redirect_uri: `${window.location.origin}/oauth/custom/callback`,
      auth_url: '',
      token_url: '',
      user_info_url: '',
      scopes: ['openid'],
    }
  }
  return {}
}

onMounted(async () => {
  await getOAuth2Setting()
  await refreshHealthCheck()
  if (OAuth2Setting.value.provider) {
    const res = await fetchGetOAuthInfo(OAuth2Setting.value.provider)
    if (res.code === 1) {
      oauthInfo.value = res.data

      if (
        oauthInfo.value.auth_type === 'oidc' &&
        oauthInfo.value.issuer &&
        oauthInfo.value.oauth_id &&
        String(oauthInfo.value.user_id || '') !== '0'
      ) {
        isBound.value = true
      } else if (
        (oauthInfo.value.auth_type === '' || oauthInfo.value.auth_type === 'oauth2') &&
        oauthInfo.value.provider &&
        oauthInfo.value.oauth_id &&
        String(oauthInfo.value.user_id || '') !== '0'
      ) {
        isBound.value = true
      } else {
        isBound.value = false
      }
    }
  }
})
</script>

<style scoped></style>
