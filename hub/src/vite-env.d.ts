/// <reference types="vite/client" />
/// <reference types="vite-plugin-pwa/client" />

interface ImportMetaEnv {
  /** GitHub Issue 预填模板，用于「加入 Hub」 */
  readonly VITE_HUB_SUBMIT_ISSUE_URL?: string
  /** Canonical / Open Graph base URL (e.g. https://hub.ech0.app) */
  readonly VITE_HUB_SITE_ORIGIN?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
