import { request, downloadFile } from '../request'

// Hello Ech0
export function fetchHelloEch0() {
  return request<App.Api.Ech0.HelloEch0>({
    url: '/hello',
    method: 'GET',
  })
}

// 执行备份
export function fetchBackup() {
  return request({
    url: '/backup',
    method: 'GET',
  })
}

// 导出备份 - 使用专门的下载函数
export function fetchExportBackup() {
  return downloadFile({
    url: '/backup/export',
    method: 'GET',
  })
}

// 导入备份
export function fetchImportBackup(file: File) {
  const formData = new FormData()
  formData.append('file', file)
  return request({
    url: '/backup/import',
    method: 'POST',
    data: formData,
  })
}

// 获取网站标题
export function fetchGetWebsiteTitle(websiteURL: string) {
  return request<string>({
    url: `/website/title?website_url=${encodeURIComponent(websiteURL)}`,
    method: 'GET',
  })
}

export interface CreateMigrationJobPayload {
  source_type: 'memos' | 'ech0_v3'
  source_version?: string
  source_payload: Record<string, any>
}

export function fetchCreateMigrationJob(data: CreateMigrationJobPayload) {
  return request({
    url: '/migration/jobs',
    method: 'POST',
    data,
  })
}

export function fetchGetMigrationJob(id: string) {
  return request({
    url: `/migration/jobs/${id}`,
    method: 'GET',
  })
}

export function fetchCancelMigrationJob(id: string) {
  return request({
    url: `/migration/jobs/${id}/cancel`,
    method: 'POST',
  })
}

export function fetchRetryFailedMigrationJob(id: string) {
  return request({
    url: `/migration/jobs/${id}/retry-failed`,
    method: 'POST',
  })
}
