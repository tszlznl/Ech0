package model

import "time"

// PageQueryDto 用于分页查询的请求数据传输对象
//
// swagger:model PageQueryDto
type PageQueryDto struct {
	Page     int    `json:"page"     form:"page"`
	PageSize int    `json:"pageSize" form:"pageSize"`
	Search   string `json:"search"   form:"search"`
}

// FileDto is the unified response for file operations.
// The Key field is the single source of truth — URLs are resolved at runtime.
//
// swagger:model FileDto
type FileDto struct {
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	Key         string `json:"key"`
	StorageType string `json:"storage_type,omitempty"`
	URL         string `json:"url"`
	ContentType string `json:"content_type,omitempty"`
	Category    string `json:"category,omitempty"`
	Size        int64  `json:"size,omitempty"`
	Width       int    `json:"width,omitempty"`
	Height      int    `json:"height,omitempty"`
}

// FileDeleteDto is the request body for deleting a file.
//
// swagger:model FileDeleteDto
type FileDeleteDto struct {
	ID string `json:"id" binding:"required"`
}

// PresignDto 用于响应 S3 预签名 URL 的请求数据传输对象
//
// swagger:model PresignDto
type PresignDto struct {
	ID          string `json:"id"`
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	Key         string `json:"key"`
	PresignURL  string `json:"presign_url"`
	FileURL     string `json:"file_url"`
}

// GetPresignURLDto 用于请求 S3 预签名 URL 的请求数据传输对象
//
// swagger:model GetPresignURLDto
type GetPresignURLDto struct {
	FileName    string `json:"file_name" binding:"required"`
	ContentType string `json:"content_type"`
	StorageType string `json:"storage_type,omitempty"`
}

// CreateExternalFileDto 用于直链文件入库请求
//
// swagger:model CreateExternalFileDto
type CreateExternalFileDto struct {
	URL         string `json:"url" binding:"required"`
	ContentType string `json:"content_type"`
	Category    string `json:"category"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Name        string `json:"name"`
}

// UpdateFileMetaDto 用于回填对象存储上传后的元信息
//
// swagger:model UpdateFileMetaDto
type UpdateFileMetaDto struct {
	Size        int64  `json:"size" binding:"required,min=0"`
	Width       *int   `json:"width,omitempty"`
	Height      *int   `json:"height,omitempty"`
	ContentType string `json:"content_type,omitempty"`
}

// FileListQueryDto 文件列表查询参数
//
// swagger:model FileListQueryDto
type FileListQueryDto struct {
	Page        int    `json:"page" form:"page"`
	PageSize    int    `json:"pageSize" form:"pageSize"`
	Search      string `json:"search" form:"search"`
	StorageType string `json:"storage_type" form:"storage_type"`
}

// FileListItemDto 文件列表项
//
// swagger:model FileListItemDto
type FileListItemDto struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Key         string    `json:"key"`
	StorageType string    `json:"storage_type"`
	URL         string    `json:"url"`
	ContentType string    `json:"content_type,omitempty"`
	Size        int64     `json:"size,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// FileListResultDto 文件列表结果
//
// swagger:model FileListResultDto
type FileListResultDto struct {
	Total int64             `json:"total"`
	Items []FileListItemDto `json:"items"`
}

// GetWebsiteTitleDto 用于请求网站标题的请求数据传输对象
//
// swagger:model GetWebsiteTitleDto
type GetWebsiteTitleDto struct {
	WebSiteURL string `json:"website_url" form:"website_url" binding:"required"`
}
