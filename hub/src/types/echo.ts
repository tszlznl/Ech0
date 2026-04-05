/** 与后端 Echo 模型对齐（见 internal/model/echo） */
export interface EchoTag {
  id: string
  name: string
}

/** 帖子正文与附件（Hub 聚合列表接口返回的字段） */
export interface EchoPost {
  id: string
  content: string
  username?: string
  created_at: string
  fav_count?: number
  tags?: EchoTag[]
  echo_files?: App.Api.Ech0.EchoFile[]
  layout?: string
  extension?: App.Api.Ech0.EchoExtension | null
  private?: boolean
  user_id?: string
}

export interface HubPostMeta {
  instanceId: string
  instanceUrl: string
}

export type EchoPostWithHub = EchoPost & { _hub: HubPostMeta }

/** 统一 API 包装 internal/model/common/result.go */
export interface ApiResult<T> {
  code: number
  msg: string
  data: T
}

export interface EchoQueryPage {
  total: number
  items: EchoPost[]
}
