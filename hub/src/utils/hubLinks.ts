/** 加入 Hub（GitHub Issue）链接，部署时可覆盖 */

const trim = (v: string | undefined) => (v ?? '').trim()

export function getHubSubmitIssueUrl(): string {
  return (
    trim(import.meta.env.VITE_HUB_SUBMIT_ISSUE_URL) ||
    'https://github.com/lin-snow/Ech0/issues/new?template=register-hub-instance.yml'
  )
}
