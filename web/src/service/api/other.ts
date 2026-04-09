import { request, downloadFile } from '../request'

// Hello Ech0
export function fetchHelloEch0() {
  return request<App.Api.Ech0.HelloEch0>({
    url: '/hello',
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

export type CheckUpdateResult = {
  current_version: string
  latest_version: string
  has_update: boolean
}

export function fetchCheckUpdate() {
  return request<CheckUpdateResult>({
    url: '/system/check-update',
    method: 'GET',
  })
}

// 获取网站标题
export function fetchGetWebsiteTitle(websiteURL: string) {
  return request<string>({
    url: `/website/title?website_url=${encodeURIComponent(websiteURL)}`,
    method: 'GET',
  })
}

export interface StartMigrationPayload {
  source_type: 'ech0_v4' | 'memos' | 'ech0_v3'
  source_payload: Record<string, unknown>
}

export interface MigrationStatusPayload extends StartMigrationPayload {
  version: number
  status: 'idle' | 'pending' | 'running' | 'success' | 'failed' | 'cancelled'
  error_message: string
  started_at?: string
  updated_at?: string
  finished_at?: string
}

export function fetchStartMigration(data: StartMigrationPayload) {
  return request({
    url: '/migration/start',
    method: 'POST',
    data,
  })
}

export function fetchGetMigrationStatus() {
  return request<MigrationStatusPayload>({
    url: '/migration/status',
    method: 'GET',
  })
}

export function fetchCancelMigration() {
  return request<MigrationStatusPayload>({
    url: '/migration/cancel',
    method: 'POST',
  })
}

export function fetchCleanupMigration() {
  return request({
    url: '/migration/cleanup',
    method: 'POST',
  })
}

export interface UploadMigrationSourceZipResponse {
  source_type: 'ech0_v4' | 'memos' | 'ech0_v3'
  tmp_dir: string
  source_payload: Record<string, unknown>
}

export function fetchUploadMigrationSourceZip(
  sourceType: UploadMigrationSourceZipResponse['source_type'],
  file: File,
) {
  const formData = new FormData()
  formData.append('source_type', sourceType)
  formData.append('file', file)
  return request<UploadMigrationSourceZipResponse>({
    url: '/migration/upload',
    method: 'POST',
    // 上传后端还会执行解压，耗时可能显著高于默认请求超时。
    timeout: 30 * 60 * 1000,
    data: formData,
  })
}
