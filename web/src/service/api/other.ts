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

// 获取网站标题
export function fetchGetWebsiteTitle(websiteURL: string) {
  return request<string>({
    url: `/website/title?website_url=${encodeURIComponent(websiteURL)}`,
    method: 'GET',
  })
}

export interface StartMigrationPayload {
  source_type: 'ech0_v4' | 'memos' | 'ech0_v3'
  source_payload: Record<string, any>
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
  source_payload: Record<string, any>
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
    data: formData,
  })
}
