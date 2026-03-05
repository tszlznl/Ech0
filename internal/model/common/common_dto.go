package model

// PageQueryDto 用于分页查询的请求数据传输对象
//
// swagger:model PageQueryDto
type PageQueryDto struct {
	Page     int    `json:"page"     form:"page"`     // 页码，从1开始
	PageSize int    `json:"pageSize" form:"pageSize"` // 每页大小
	Search   string `json:"search"   form:"search"`   // 用于搜索的关键字
}

// FileDto 用于文件相关的请求数据传输对象
//
// swagger:model FileDto
type FileDto struct {
	// 文件的 URL 地址
	URL string `json:"url" binding:"required"`
	// 可直接访问地址（前端渲染应优先使用）
	AccessURL string `json:"access_url,omitempty"`
	// 文件来源，如 local/s3/url
	Source string `json:"source" binding:"required"`
	// 对象存储的 Key, 用于删除 S3/R2 上的文件
	ObjectKey string `json:"object_key"`
	// MIME 类型
	ContentType string `json:"content_type,omitempty"`
	// 文件类别，如 image/audio
	Category string `json:"category,omitempty"`
	// 分类元数据（不同类别填充不同字段）
	Metadata FileMetadataDto `json:"metadata,omitempty"`
	// 图片宽度（非图片可为空）
	Width int `json:"width"`
	// 图片高度（非图片可为空）
	Height int `json:"height"`
}

type FileMetadataDto struct {
	Image    *ImageMetadataDto    `json:"image,omitempty"`
	Video    *VideoMetadataDto    `json:"video,omitempty"`
	Audio    *AudioMetadataDto    `json:"audio,omitempty"`
	PDF      *PDFMetadataDto      `json:"pdf,omitempty"`
	Markdown *MarkdownMetadataDto `json:"markdown,omitempty"`
}

type ImageMetadataDto struct {
	Width  int `json:"width,omitempty"`
	Height int `json:"height,omitempty"`
}

type VideoMetadataDto struct {
	Width      int `json:"width,omitempty"`
	Height     int `json:"height,omitempty"`
	DurationMs int `json:"duration_ms,omitempty"`
}

type AudioMetadataDto struct {
	DurationMs int `json:"duration_ms,omitempty"`
}

type PDFMetadataDto struct {
	Pages int `json:"pages,omitempty"`
}

type MarkdownMetadataDto struct {
	WordCount int `json:"word_count,omitempty"`
}

// ImageDto 兼容旧版图片 DTO（建议使用 FileDto）
//
// swagger:model ImageDto
type ImageDto struct {
	URL       string `json:"url"        binding:"required"`
	SOURCE    string `json:"source"     binding:"required"`
	ObjectKey string `json:"object_key"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
}

// PresignDto 用于响应 S3 预签名 URL 的请求数据传输对象
//
// swagger:model PresignDto
type PresignDto struct {
	FileName    string `json:"file_name"` // 原始文件名
	ContentType string `json:"content_type"`
	ObjectKey   string `json:"object_key"`  // 预签名的对象存储 Key
	PresignURL  string `json:"presign_url"` // 预签名 URL
	FileURL     string `json:"file_url"`    // 文件访问 URL,用于回显
}

// GetPresignURLDto 用于请求 S3 预签名 URL 的请求数据传输对象
//
// swagger:model GetPresignURLDto
type GetPresignURLDto struct {
	FileName    string `json:"file_name"    binding:"required"` // 原始文件名
	ContentType string `json:"content_type"`                    // 文件的 MIME 类型
}

// GetWebsiteTitleDto 用于请求网站标题的请求数据传输对象
//
// swagger:model GetWebsiteTitleDto
type GetWebsiteTitleDto struct {
	WebSiteURL string `json:"website_url" form:"website_url" binding:"required"` // 网站URL
}
