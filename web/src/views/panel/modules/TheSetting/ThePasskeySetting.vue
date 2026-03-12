<template>
  <PanelCard>
    <div class="w-full">
      <div class="flex flex-row items-center justify-between mb-3">
        <h1 class="text-[var(--color-text-primary)] font-bold text-lg">Passkey</h1>
        <BaseEditCapsule
          :editing="passkeyEditMode"
          apply-title="应用"
          cancel-title="取消"
          edit-title="编辑"
          @apply="handleUpdatePasskeySetting"
          @toggle="passkeyEditMode = !passkeyEditMode"
        />
      </div>

      <div class="text-[var(--color-text-muted)] text-sm mb-3">
        使用 Passkey（WebAuthn）可在不同设备上无密码登录
      </div>

      <div class="mb-3 border border-dashed border-[var(--color-border-strong)] rounded-md p-3">
        <h2 class="text-[var(--color-text-primary)] font-semibold mb-2">配置健康检查</h2>
        <div class="flex flex-col sm:flex-row sm:flex-wrap gap-2 text-sm">
          <div class="flex items-center gap-2">
            <span class="text-[var(--color-text-secondary)]">Passkey就绪:</span>
            <span
              class="px-2 py-0.5 rounded-md"
              :class="
                passkeyRuntimeStatus?.passkey_ready
                  ? 'bg-green-500/15 text-green-500'
                  : 'bg-yellow-500/15 text-yellow-500'
              "
            >
              {{ passkeyRuntimeStatus?.passkey_ready ? '已就绪' : '未就绪' }}
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

      <div class="mb-3 border border-dashed border-[var(--color-border-strong)] rounded-md p-3">
        <h2 class="text-[var(--color-text-primary)] font-semibold mb-2">Passkey 安全边界</h2>
        <div
          class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 h-10"
        >
          <h3 class="font-semibold w-40 shrink-0">WebAuthn RP ID:</h3>
          <span v-if="!passkeyEditMode" class="truncate max-w-80 inline-block align-middle">
            {{ passkeySetting.webauthn_rp_id || '暂无' }}
          </span>
          <BaseInput
            v-else
            v-model="passkeySetting.webauthn_rp_id"
            type="text"
            placeholder="例如：example.com"
            class="w-full py-1!"
          />
        </div>
        <div
          class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 h-10"
        >
          <h3 class="font-semibold w-40 shrink-0">WebAuthn Origins:</h3>
          <span v-if="!passkeyEditMode" class="truncate max-w-80 inline-block align-middle">
            {{
              passkeySetting.webauthn_allowed_origins.length === 0
                ? '暂无'
                : passkeySetting.webauthn_allowed_origins.join(', ')
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
      </div>

      <!-- 绑定 -->
      <div class="flex items-center justify-start gap-2 mb-4">
        <div class="w-50">
          <BaseInput
            v-model="newDeviceName"
            type="text"
            placeholder="设备名称（可选）"
            class="py-1 text-sm"
          />
        </div>
        <BaseButton
          class="rounded-md px-3 w-14 h-9 text-sm flex items-center justify-center"
          :disabled="busy || !supported"
          @click="handleBind"
        >
          绑定
        </BaseButton>
      </div>

      <div v-if="!supported" class="text-[var(--color-text-muted)] text-sm mb-3">
        当前浏览器不支持 Passkey / WebAuthn。
      </div>

      <!-- 多设备管理 -->
      <div class="text-[var(--color-text-muted)] font-semibold mb-2">已绑定设备</div>
      <div v-if="devices.length === 0" class="text-[var(--color-text-muted)] text-sm">暂无设备</div>
      <div
        v-else
        class="mt-2 overflow-x-auto border border-[var(--color-border-subtle)] rounded-lg"
      >
        <table class="min-w-full divide-y divide-[var(--color-border-subtle)]">
          <thead>
            <tr class="bg-[var(--color-bg-surface)] opacity-70">
              <th
                class="px-3 py-2 text-left text-sm font-semibold text-[var(--color-text-primary)]"
              >
                设备名称
              </th>
              <th
                class="px-3 py-2 text-left text-sm font-semibold text-[var(--color-text-primary)]"
              >
                AAGUID
              </th>
              <th
                class="px-3 py-2 text-left text-sm font-semibold text-[var(--color-text-primary)]"
              >
                时间
              </th>
              <th
                class="px-3 py-2 text-right text-sm font-semibold text-[var(--color-text-primary)]"
              >
                操作
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-[var(--color-border-subtle)] text-nowrap">
            <tr v-for="d in devices" :key="d.id">
              <td class="px-3 py-2 text-sm text-[var(--color-text-primary)] font-semibold">
                {{ d.device_name || 'Passkey' }}
              </td>
              <td class="px-3 py-2 text-sm text-[var(--color-text-secondary)]">
                {{ d.aaguid || '未知' }}
              </td>
              <td class="px-3 py-2 text-xs text-[var(--color-text-secondary)]">
                <div>最近使用：{{ formatTime(d.last_used_at) }}</div>
                <div>创建：{{ formatTime(d.created_at) }}</div>
              </td>
              <td class="px-3 py-2 text-right">
                <div class="flex flex-row items-center justify-end gap-2">
                  <BaseButton
                    class="rounded-md"
                    :disabled="busy"
                    @click="promptRename(d)"
                    title="改名"
                  >
                    <Rename class="w-5 h-5" />
                  </BaseButton>
                  <BaseButton
                    class="rounded-md"
                    :disabled="busy"
                    @click="handleDelete(d.id)"
                    title="删除"
                  >
                    <Trashbin class="w-5 h-5" />
                  </BaseButton>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </PanelCard>
</template>
<script setup lang="ts">
import { watch, onMounted, ref } from 'vue'
import PanelCard from '@/layout/PanelCard.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseEditCapsule from '@/components/common/BaseEditCapsule.vue'
import Trashbin from '@/components/icons/trashbin.vue'
import Rename from '@/components/icons/rename.vue'
import {
  fetchGetPasskeySettings,
  fetchGetPasskeyStatus,
  fetchDeletePasskeyDevice,
  fetchPasskeyDevices,
  fetchPasskeyRegisterBegin,
  fetchPasskeyRegisterFinish,
  fetchUpdatePasskeySettings,
  fetchUpdatePasskeyDeviceName,
} from '@/service/api'
import { theToast } from '@/utils/toast'
import { useBaseDialog } from '@/composables/useBaseDialog'
import { base64urlToUint8Array, uint8ArrayToBase64url } from '@/utils/other'
const { openConfirm } = useBaseDialog()

const supported = !!(window.PublicKeyCredential && navigator.credentials)
const busy = ref(false)
const newDeviceName = ref<string>('My Passkey')
const devices = ref<App.Api.Auth.PasskeyDevice[]>([])
const passkeyEditMode = ref(false)
const passkeySetting = ref<App.Api.Setting.PasskeySetting>({
  webauthn_rp_id: '',
  webauthn_allowed_origins: [],
})
const passkeyRuntimeStatus = ref<App.Api.Setting.PasskeyStatus | null>(null)
const webauthnOriginsString = ref('')
const missingBoundaryItems = ref<string[]>([])

type Base64urlString = string

type CredentialDescriptorJSON = {
  type: PublicKeyCredentialType
  id: Base64urlString
  transports?: AuthenticatorTransport[]
}

type UserEntityJSON = {
  id: Base64urlString
  name: string
  displayName: string
}

type CreationOptionsJSON = Omit<
  PublicKeyCredentialCreationOptions,
  'challenge' | 'user' | 'excludeCredentials'
> & {
  challenge: Base64urlString
  user: UserEntityJSON
  excludeCredentials?: CredentialDescriptorJSON[]
}

const parseList = (input: string) =>
  input
    .split(',')
    .map((s) => s.trim())
    .filter((s) => s.length > 0)

// 断言服务端返回的 publicKey 合法
function assertCreationOptionsJSON(raw: unknown): CreationOptionsJSON {
  if (!raw || typeof raw !== 'object') throw new Error('服务端返回的 publicKey 不合法')
  return raw as CreationOptionsJSON
}

// 标准化服务端返回的 publicKey
function normalizeCreationOptions(raw: unknown): PublicKeyCredentialCreationOptions {
  const o = assertCreationOptionsJSON(raw)
  const { challenge, user, excludeCredentials, ...rest } = o
  const exclude = Array.isArray(excludeCredentials)
    ? excludeCredentials.map((c) => ({
        ...c,
        id: base64urlToUint8Array(c.id) as BufferSource,
      }))
    : undefined

  return {
    ...rest,
    challenge: base64urlToUint8Array(challenge) as BufferSource,
    user: {
      ...user,
      id: base64urlToUint8Array(user.id) as BufferSource,
    },
    ...(exclude ? { excludeCredentials: exclude } : {}),
  } as PublicKeyCredentialCreationOptions
}

// 将 PublicKeyCredential 转换为 JSON
function credentialToJSON(cred: PublicKeyCredential) {
  if (!cred) return null
  const obj: Record<string, unknown> = {
    id: cred.id,
    rawId: uint8ArrayToBase64url(cred.rawId),
    type: cred.type,
    clientExtensionResults: cred.getClientExtensionResults?.() ?? {},
  }

  const response: Record<string, unknown> = {}
  response.clientDataJSON = uint8ArrayToBase64url(cred.response.clientDataJSON)

  // 注册（attestation）
  if ('attestationObject' in cred.response) {
    const r = cred.response as AuthenticatorAttestationResponse
    response.attestationObject = uint8ArrayToBase64url(r.attestationObject)
  }

  // 登录（assertion）——这里暂时不会用到，但保持通用
  if ('authenticatorData' in cred.response) {
    const r = cred.response as AuthenticatorAssertionResponse
    response.authenticatorData = uint8ArrayToBase64url(r.authenticatorData)
    response.signature = uint8ArrayToBase64url(r.signature)
    if (r.userHandle && r.userHandle.byteLength > 0) {
      response.userHandle = uint8ArrayToBase64url(r.userHandle)
    }
  }

  obj.response = response
  return obj
}

// 格式化时间
function formatTime(v: string) {
  if (!v) return '暂无'
  const d = new Date(v)
  if (Number.isNaN(d.getTime())) return v
  return d.toLocaleString()
}

// 刷新设备列表
async function refresh() {
  const res = await fetchPasskeyDevices()
  if (res.code === 1) devices.value = res.data ?? []
}

async function getPasskeySetting() {
  const res = await fetchGetPasskeySettings()
  if (res.code === 1) {
    passkeySetting.value = res.data
    webauthnOriginsString.value = (res.data.webauthn_allowed_origins || []).join(', ')
  }
}

async function refreshHealthCheck() {
  const statusRes = await fetchGetPasskeyStatus()
  if (statusRes.code === 1) {
    passkeyRuntimeStatus.value = statusRes.data
  }
  const missing: string[] = []
  if (!passkeySetting.value.webauthn_rp_id) {
    missing.push('WebAuthn RP ID')
  }
  if ((passkeySetting.value.webauthn_allowed_origins || []).length === 0) {
    missing.push('WebAuthn Origins')
  }
  missingBoundaryItems.value = missing
}

async function handleUpdatePasskeySetting() {
  passkeySetting.value.webauthn_allowed_origins = parseList(webauthnOriginsString.value)
  if (passkeySetting.value.webauthn_allowed_origins.some((u) => !/^https?:\/\//.test(u))) {
    theToast.error('WebAuthn Origins 需为 http/https URL')
    return
  }
  const res = await fetchUpdatePasskeySettings(passkeySetting.value)
  if (res.code === 1) {
    theToast.success(res.msg)
    passkeyEditMode.value = false
    await getPasskeySetting()
    await refreshHealthCheck()
  }
}

function handleAutoFillBoundary() {
  const currentOrigin = window.location.origin
  const currentHost = window.location.hostname
  if (!passkeySetting.value.webauthn_rp_id) {
    passkeySetting.value.webauthn_rp_id = currentHost
  }
  if (!passkeySetting.value.webauthn_allowed_origins?.length) {
    passkeySetting.value.webauthn_allowed_origins = [currentOrigin]
  }
  webauthnOriginsString.value = passkeySetting.value.webauthn_allowed_origins.join(', ')
  passkeyEditMode.value = true
  void refreshHealthCheck()
  theToast.success('已填充推荐配置，请点击“应用”保存')
}

// 绑定设备
async function handleBind() {
  if (!supported) return
  busy.value = true
  try {
    const begin = await fetchPasskeyRegisterBegin(newDeviceName.value || 'Passkey')
    if (begin.code !== 1) return

    const options = normalizeCreationOptions(begin.data.publicKey)
    const created = await navigator.credentials.create({ publicKey: options })
    if (!created) throw new Error('创建凭证失败')
    const cred = created as PublicKeyCredential

    const finish = await fetchPasskeyRegisterFinish(begin.data.nonce, credentialToJSON(cred))
    if (finish.code !== 1) return

    theToast.success('绑定成功')
    await refresh()
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : '绑定失败'
    theToast.error(msg)
  } finally {
    busy.value = false
  }
}

// 删除设备
async function handleDelete(id: string) {
  openConfirm({
    title: '确定要删除该设备吗？',
    description: '删除后该设备将无法登录，请谨慎操作',
    onConfirm: async () => {
      busy.value = true
      try {
        const res = await fetchDeletePasskeyDevice(id)
        if (res.code !== 1) return
        theToast.success('已删除')
        await refresh()
      } finally {
        busy.value = false
      }
    },
  })
}

// 改名
async function promptRename(d: App.Api.Auth.PasskeyDevice) {
  const name = window.prompt('新的设备名称', d.device_name || 'Passkey')
  if (!name) return
  busy.value = true
  try {
    const res = await fetchUpdatePasskeyDeviceName(d.id, name)
    if (res.code !== 1) return
    theToast.success('已更新')
    await refresh()
  } finally {
    busy.value = false
  }
}

watch(
  () => passkeySetting.value,
  (v) => {
    webauthnOriginsString.value = (v.webauthn_allowed_origins || []).join(', ')
  },
  { immediate: true, deep: true },
)

onMounted(async () => {
  await getPasskeySetting()
  await refreshHealthCheck()
  await refresh()
})
</script>
