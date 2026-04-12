import { request } from '../request'

export function fetchGetVisitorStats() {
  return request<App.Api.Dashboard.VisitorDayStat[]>({
    url: '/system/visitor-stats',
    method: 'GET',
  })
}
