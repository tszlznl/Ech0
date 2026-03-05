package storage

import (
	"strings"
	"time"
)

type Category string

const (
	CategoryImage    Category = "image"
	CategoryVideo    Category = "video"
	CategoryAudio    Category = "audio"
	CategoryPDF      Category = "pdf"
	CategoryMarkdown Category = "markdown"
	CategoryFile     Category = "file"
)

type Source string

const (
	SourceLocal Source = "local"
	SourceS3    Source = "s3"
	SourceURL   Source = "url"
)

type ImageMetadata struct {
	Width  int `json:"width,omitempty"`
	Height int `json:"height,omitempty"`
}

type VideoMetadata struct {
	Width      int `json:"width,omitempty"`
	Height     int `json:"height,omitempty"`
	DurationMs int `json:"duration_ms,omitempty"`
}

type AudioMetadata struct {
	DurationMs int `json:"duration_ms,omitempty"`
}

type PDFMetadata struct {
	Pages int `json:"pages,omitempty"`
}

type MarkdownMetadata struct {
	WordCount int `json:"word_count,omitempty"`
}

type FileMetadata struct {
	Image    *ImageMetadata    `json:"image,omitempty"`
	Video    *VideoMetadata    `json:"video,omitempty"`
	Audio    *AudioMetadata    `json:"audio,omitempty"`
	PDF      *PDFMetadata      `json:"pdf,omitempty"`
	Markdown *MarkdownMetadata `json:"markdown,omitempty"`
}

type FileObject struct {
	URL         string       `json:"url"`
	Source      Source       `json:"source"`
	ObjectKey   string       `json:"object_key,omitempty"`
	ContentType string       `json:"content_type,omitempty"`
	Category    Category     `json:"category,omitempty"`
	Metadata    FileMetadata `json:"metadata,omitempty"`
}

type SaveRequest struct {
	UserID      uint
	FileName    string
	ContentType string
	Category    Category
	Reader      ReadSeekCloser
}

type DeleteRequest struct {
	URL       string
	Source    Source
	ObjectKey string
	Category  Category
}

type PresignRequest struct {
	UserID      uint
	FileName    string
	ContentType string
	Method      string
	Expiry      time.Duration
	Category    Category
}

type PresignResponse struct {
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	ObjectKey   string `json:"object_key"`
	PresignURL  string `json:"presign_url"`
	FileURL     string `json:"file_url"`
}

type ReadSeekCloser interface {
	Read(p []byte) (n int, err error)
	Seek(offset int64, whence int) (int64, error)
	Close() error
}

func NormalizeCategory(raw string) Category {
	switch Category(strings.ToLower(strings.TrimSpace(raw))) {
	case CategoryImage:
		return CategoryImage
	case CategoryVideo:
		return CategoryVideo
	case CategoryAudio:
		return CategoryAudio
	case CategoryPDF:
		return CategoryPDF
	case CategoryMarkdown:
		return CategoryMarkdown
	default:
		return CategoryFile
	}
}

func (c Category) IsImageLike() bool {
	return c == CategoryImage
}

