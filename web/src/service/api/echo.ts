import { request, requestWithDirectUrlAndData } from '../request'

// 分页获取Echos
export async function fetchGetEchosByPage(searchParams: App.Api.Ech0.ParamsByPagination) {
  return request<App.Api.Ech0.PaginationResult>({
    url: `/echo/page`,
    method: 'POST',
    data: searchParams,
  })
}

// 添加Echo
export function fetchAddEcho(echoToAdd: App.Api.Ech0.EchoToAdd) {
  return request({
    url: `/echo`,
    method: 'POST',
    data: echoToAdd,
  })
}

// 删除Echo
export function fetchDeleteEcho(echoId: string) {
  return request({
    url: `/echo/${echoId}`,
    method: 'DELETE',
  })
}

// 更新Echo
export function fetchUpdateEcho(echo: App.Api.Ech0.EchoToUpdate) {
  return request({
    url: `/echo`,
    method: 'PUT',
    data: echo,
  })
}

// 点赞Echo
export function fetchLikeEcho(echoId: string) {
  return request({
    url: `/echo/like/${echoId}`,
    method: 'PUT',
  })
}

// 获取Echo详情
export async function fetchGetEchoById(echoId: string) {
  return request<App.Api.Ech0.Echo>({
    url: `/echo/${echoId}`,
    method: 'GET',
  })
}

// 获取status
export function fetchGetStatus() {
  return request<App.Api.Ech0.Status>({
    url: `/status`,
    method: 'GET',
  })
}

// 获取一个月内的热力图
export function fetchGetHeatMap() {
  return request<App.Api.Ech0.HeatMap>({
    url: `/heatmap`,
    method: 'GET',
  })
}

// 获取Github仓库数据
export function fetchGetGithubRepo(githubRepo: { owner: string; repo: string }) {
  return requestWithDirectUrlAndData<App.Api.Ech0.GithubCardData>({
    dirrectUrlAndData: `https://api.github.com/repos/${githubRepo.owner}/${githubRepo.repo}`,
    url: `/github`,
    method: 'GET',
  })
}

// 获取标签列表
export function fetchGetTags() {
  return request<App.Api.Ech0.Tag[]>({
    url: `/tags`,
    method: 'GET',
  })
}

// 删除某个标签
export function fetchDeleteTagById(tagId: string) {
  return request({
    url: `/tag/${tagId}`,
    method: 'DELETE',
  })
}

// 根据标签查询Echos（支持分页）
export async function fetchGetEchosByTagId(tagId: string, searchParams: App.Api.Ech0.ParamsByPagination) {
  return request<App.Api.Ech0.PaginationResult>({
    url: `/echo/tag/${tagId}?page=${searchParams.page}&pageSize=${searchParams.pageSize}&search=${searchParams.search || ''}`,
    method: 'GET',
  })
}
