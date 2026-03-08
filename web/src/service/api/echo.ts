import { request, requestWithDirectUrlAndData } from '../request'

type RawEchoFile = {
  file_id?: string
  sort_order?: number
  file?: {
    id?: string
    key?: string
    url?: string
    storage_type?: string
    width?: number
    height?: number
  }
}

type RawEcho = App.Api.Ech0.Echo & {
  echo_files?: RawEchoFile[]
}

function normalizeEcho(rawEcho: RawEcho): App.Api.Ech0.Echo {
  const images = (rawEcho.echo_files || []).map((item) => {
    const file: NonNullable<RawEchoFile['file']> = item.file || {}
    const fileUrl = String(file.url || '')
    return {
      id: String(file.id || item.file_id || ''),
      echo_id: String(rawEcho.id || ''),
      url: fileUrl,
      image_source: file.storage_type === 'object' ? 's3' : 'local',
      object_key: String(file.key || ''),
      width: file.width,
      height: file.height,
    }
  })
  return {
    ...rawEcho,
    images,
  }
}

function normalizePaginationResult(result: App.Api.Ech0.PaginationResult) {
  return {
    ...result,
    items: (result.items || []).map((item) => normalizeEcho(item as RawEcho)),
  }
}

function buildEchoPayload(
  echoToAddOrUpdate: App.Api.Ech0.EchoToAdd | App.Api.Ech0.EchoToUpdate,
) {
  const images = echoToAddOrUpdate.images || []
  const echoFiles = images
    .filter((img) => img.id)
    .map((img, index) => ({
      file_id: String(img.id),
      sort_order: index,
    }))
  return {
    ...echoToAddOrUpdate,
    echo_files: echoFiles,
  }
}

// 分页获取Echos
export async function fetchGetEchosByPage(searchParams: App.Api.Ech0.ParamsByPagination) {
  const res = await request<App.Api.Ech0.PaginationResult>({
    url: `/echo/page`,
    method: 'POST',
    data: searchParams,
  })
  if (res.code === 1) {
    res.data = normalizePaginationResult(res.data)
  }
  return res
}

// 添加Echo
export function fetchAddEcho(echoToAdd: App.Api.Ech0.EchoToAdd) {
  return request({
    url: `/echo`,
    method: 'POST',
    data: buildEchoPayload(echoToAdd),
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
    data: buildEchoPayload(echo),
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
  const res = await request<App.Api.Ech0.Echo>({
    url: `/echo/${echoId}`,
    method: 'GET',
  })
  if (res.code === 1 && res.data) {
    res.data = normalizeEcho(res.data as RawEcho)
  }
  return res
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

// 获取预签名URL
export function fetchGetPresignedUrl(fileName: string, contentType?: string) {
  return request<App.Api.Ech0.PresignResult>({
    url: `/files/presign`,
    method: 'PUT',
    data: {
      file_name: fileName,
      content_type: contentType,
    },
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
  const res = await request<App.Api.Ech0.PaginationResult>({
    url: `/echo/tag/${tagId}?page=${searchParams.page}&pageSize=${searchParams.pageSize}&search=${searchParams.search || ''}`,
    method: 'GET',
  })
  if (res.code === 1) {
    res.data = normalizePaginationResult(res.data)
  }
  return res
}
