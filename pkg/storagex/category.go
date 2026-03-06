package storagex

import "strings"

type Category string

const (
	CategoryImage    Category = "image"
	CategoryVideo    Category = "video"
	CategoryAudio    Category = "audio"
	CategoryPDF      Category = "pdf"
	CategoryMarkdown Category = "markdown"
	CategoryFile     Category = "file"
)

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

// CategoryDir maps a category to its virtual directory name.
func CategoryDir(c Category) string {
	switch c {
	case CategoryImage:
		return "images"
	case CategoryVideo:
		return "videos"
	case CategoryAudio:
		return "audios"
	case CategoryPDF:
		return "documents"
	case CategoryMarkdown:
		return "documents"
	default:
		return "files"
	}
}
