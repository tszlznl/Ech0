import { request } from '../request'
import { FILE_CATEGORY, FILE_STORAGE_TYPE } from '@/constants/file'

// 上传文件
export function fetchUploadFile(
  file: File,
  storageType: App.Api.File.StorageType = FILE_STORAGE_TYPE.LOCAL,
  category: App.Api.File.Category = FILE_CATEGORY.IMAGE,
) {
  const formData = new FormData()
  formData.append('file', file)
  formData.append('storage_type', storageType)
  formData.append('category', category)

  return request<App.Api.File.FileDto>({
    url: `/files/upload`,
    method: 'POST',
    data: formData,
  })
}

// 创建外链文件记录（不上传二进制）
export function fetchCreateExternalFile(dto: App.Api.File.CreateExternalFileDto) {
  return request<App.Api.File.FileDto>({
    url: `/files/external`,
    method: 'POST',
    data: dto,
  })
}

// 删除文件
export function fetchDeleteFile(file: App.Api.File.FileDeleteDto) {
  return request({
    url: `/file/${file.id}`,
    method: 'DELETE',
  })
}

// 按ID获取文件详情
export function fetchGetFileById(id: string) {
  return request<App.Api.File.FileDto>({
    url: `/file/${id}`,
    method: 'GET',
  })
}

// 获取预签名URL（对象存储）
export function fetchGetPresignedUrl(
  fileName: string,
  contentType?: string,
  storageType: App.Api.File.StorageType = FILE_STORAGE_TYPE.OBJECT,
) {
  return request<App.Api.Ech0.PresignResult>({
    url: `/files/presign`,
    method: 'PUT',
    data: {
      file_name: fileName,
      content_type: contentType,
      storage_type: storageType,
    },
  })
}

