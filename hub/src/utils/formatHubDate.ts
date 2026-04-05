/** Hub 卡片底部时间展示（不依赖 web/utils/other 的 i18n 相对时间链） */
export function formatHubDate(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(d)
}
