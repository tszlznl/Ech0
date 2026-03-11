import { request } from '../request'

export function fetchGetInitStatus() {
  return request<App.Api.Init.Status>({
    url: '/init/status',
    method: 'GET',
  })
}

export function fetchInitOwner(payload: App.Api.Auth.SignupParams) {
  return request({
    url: '/init/owner',
    method: 'POST',
    data: payload,
  })
}
