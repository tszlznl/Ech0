import { request, requestWithDirectUrl } from '../request'

// 获取Connect列表
export function fetchGetConnectList() {
  return request<App.Api.Connect.Connected[]>({
    url: '/connect/list',
    method: 'GET',
  })
}

// 获取Connect详情 (直接根据URL获取，不需要request的url)
export function fetchGetConnect(connectUrl: string, silentError = false) {
  return requestWithDirectUrl<App.Api.Connect.Connect>({
    dirrectUrl: `${connectUrl}/api/connect`,
    url: '/',
    method: 'GET',
    silentError,
  })
}

// 获取所有Connect详情
export function fetchGetAllConnectInfo() {
  return request<App.Api.Connect.Connect[]>({
    url: '/connects/info',
    method: 'GET',
  })
}

// 添加Connect
export function fetchAddConnect(connectUrl: string) {
  return request<App.Api.Connect.Connected>({
    url: '/addConnect',
    method: 'POST',
    data: {
      connect_url: connectUrl,
    },
  })
}

// 删除Connect
export function fetchDeleteConnect(id: string) {
  return request<App.Api.Connect.Connected>({
    url: `/delConnect/${id}`,
    method: 'DELETE',
  })
}
