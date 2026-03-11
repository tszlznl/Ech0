import { defineStore } from 'pinia'
import { ref } from 'vue'
import {
  fetchGetSettings,
  fetchGetCommentSettings,
  fetchGetS3Settings,
  fetchGetOAuth2Settings,
  fetchGetAllWebhooks,
  fetchListAccessTokens,
  fetchGetBackupScheduleSetting,
  fetchGetAgentSettings,
  fetchGetAgentInfo,
  fetchHelloEch0,
} from '@/service/api'
import { CommentProvider, S3Provider, OAuth2Provider, AgentProvider } from '@/enums/enums'

export const useSettingStore = defineStore('settingStore', () => {
  /**
   * State
   */
  const SystemSetting = ref<App.Api.Setting.SystemSetting>({
    site_title: import.meta.env.VITE_APP_TITLE,
    server_logo: '/Ech0.svg',
    server_name: import.meta.env.VITE_APP_NAME,
    server_url: '',
    allow_register: true,
    ICP_number: '',
    footer_content: '',
    footer_link: '',
    meting_api: '',
    custom_css: '',
    custom_js: '',
  })
  const CommentSetting = ref<App.Api.Setting.CommentSetting>({
    enable_comment: false,
    provider: CommentProvider.TWIKOO,
    comment_api: '',
  })
  const S3Setting = ref<App.Api.Setting.S3Setting>({
    enable: false,
    provider: S3Provider.AWS,
    endpoint: '',
    access_key: '',
    secret_key: '',
    bucket_name: '',
    region: '',
    use_ssl: false,
    cdn_url: '',
    path_prefix: '',
    public_read: true,
  })
  const OAuth2Setting = ref<App.Api.Setting.OAuth2Setting>({
    enable: false,
    provider: OAuth2Provider.GITHUB,
    client_id: '',
    client_secret: '',
    redirect_uri: '',
    scopes: [],
    auth_url: '',
    token_url: '',
    user_info_url: '',

    is_oidc: false,
    issuer: '',
    jwks_url: '',
  })
  const Webhooks = ref<App.Api.Setting.Webhook[]>([])
  const AccessTokens = ref<App.Api.Setting.AccessToken[]>([])
  const BackupSchedule = ref<App.Api.Setting.BackupSchedule>({
    enable: false,
    cron_expression: '0 2 * * 0',
  })
  const AgentSetting = ref<App.Api.Setting.AgentSetting>({
    enable: false,
    provider: AgentProvider.OPENAI,
    model: '',
    api_key: '',
    prompt: '',
    base_url: '',
  })
  const hello = ref<App.Api.Ech0.HelloEch0>()
  const loading = ref<boolean>(true)

  /**
   * Actions
   */
  const getSystemSetting = async () => {
    await fetchGetSettings().then((res) => {
      if (res.code === 1) {
        SystemSetting.value = res.data
        loading.value = false
      }
    })
  }

  const getCommentSetting = async () => {
    fetchGetCommentSettings().then((res) => {
      if (res.code === 1) {
        CommentSetting.value = res.data
      }
    })
  }

  const getS3Setting = async () => {
    fetchGetS3Settings().then((res) => {
      if (res.code === 1) {
        S3Setting.value = res.data
      }
    })
  }

  const getOAuth2Setting = async () => {
    const res = await fetchGetOAuth2Settings()
    if (res.code === 1) {
      OAuth2Setting.value = res.data
    }
    return OAuth2Setting.value
  }

  const getAllWebhooks = async () => {
    const res = await fetchGetAllWebhooks()
    if (res.code === 1) {
      if (res.data) {
        Webhooks.value = res.data
      } else {
        Webhooks.value = []
      }
    }
  }

  const getAllAccessTokens = async () => {
    const res = await fetchListAccessTokens()
    if (res.code === 1) {
      if (res.data) {
        AccessTokens.value = res.data
      } else {
        AccessTokens.value = []
      }
    }
  }

  const getBackupSchedule = async () => {
    const res = await fetchGetBackupScheduleSetting()
    if (res.code === 1) {
      BackupSchedule.value = res.data
    }
  }

  const getHelloEch0 = async () => {
    const res = await fetchHelloEch0()
    if (res.code === 1) {
      hello.value = res.data
    }
  }

  const getAgentSetting = async () => {
    const res = await fetchGetAgentSettings()
    if (res.code === 1) {
      AgentSetting.value = res.data
    }
  }

  const getAgentInfo = async () => {
    const res = await fetchGetAgentInfo()
    if (res.code === 1) {
      AgentSetting.value.enable = res.data.enable
      AgentSetting.value.model = res.data.model
      AgentSetting.value.provider = res.data.provider
    }
  }

  const init = async () => {
    await getSystemSetting()
    getCommentSetting()
    getS3Setting()
    getAgentInfo()
    getHelloEch0()
  }

  return {
    SystemSetting,
    CommentSetting,
    S3Setting,
    OAuth2Setting,
    Webhooks,
    AccessTokens,
    BackupSchedule,
    AgentSetting,
    hello,
    loading,

    getAllAccessTokens,
    getSystemSetting,
    getCommentSetting,
    getS3Setting,
    getOAuth2Setting,
    getAllWebhooks,
    getHelloEch0,
    getBackupSchedule,
    getAgentSetting,
    getAgentInfo,
    init,
  }
})
