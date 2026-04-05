/// <reference types="vite/client" />
/// <reference types="vite-plugin-pwa/client" />

interface ImportMetaEnv {
  /** GitHub Issue 预填模板，用于「加入 Hub」 */
  readonly VITE_HUB_SUBMIT_ISSUE_URL?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
