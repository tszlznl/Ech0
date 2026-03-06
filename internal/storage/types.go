package storage

import (
	stgx "github.com/lin-snow/ech0/pkg/storagex"
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

func NormalizeCategory(raw string) Category {
	return Category(stgx.NormalizeCategory(raw))
}

func (c Category) IsImageLike() bool {
	return c == CategoryImage
}
