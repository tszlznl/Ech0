package storage

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// Category classifies uploaded files.
type Category string

const (
	CategoryImage    Category = "image"
	CategoryVideo    Category = "video"
	CategoryAudio    Category = "audio"
	CategoryPDF      Category = "pdf"
	CategoryMarkdown Category = "markdown"
	CategoryFile     Category = "file"
)

// NormalizeCategory maps an arbitrary string to a known Category.
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

// CategoryDir maps a category to its storage directory prefix.
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

// IsImageLike reports whether the category represents visual media.
func (c Category) IsImageLike() bool {
	return c == CategoryImage
}

// ---------------------------------------------------------------------------
// Key generation
// ---------------------------------------------------------------------------

// KeyGenerator produces a VireFS key for a new upload.
type KeyGenerator interface {
	GenerateKey(category Category, userID uint, originalFilename string) (string, error)
}

// RandomKeyGenerator creates keys like images/1_1700000000_abc123.png.
type RandomKeyGenerator struct {
	RandSource  io.Reader
	SuffixBytes int
	Now         func() time.Time
}

func NewRandomKeyGenerator() *RandomKeyGenerator {
	return &RandomKeyGenerator{
		RandSource:  rand.Reader,
		SuffixBytes: 4,
		Now:         time.Now,
	}
}

func (g *RandomKeyGenerator) GenerateKey(category Category, userID uint, originalFilename string) (string, error) {
	ext := strings.ToLower(filepath.Ext(strings.TrimSpace(originalFilename)))
	if ext == "" {
		ext = ".bin"
	}
	size := g.SuffixBytes
	if size <= 0 {
		size = 4
	}
	src := g.RandSource
	if src == nil {
		src = rand.Reader
	}
	randPart := make([]byte, size)
	if _, err := io.ReadFull(src, randPart); err != nil {
		return "", err
	}
	now := g.Now()
	dir := CategoryDir(category)
	filename := fmt.Sprintf("%d_%d_%s%s", userID, now.UTC().Unix(), hex.EncodeToString(randPart), ext)
	return JoinKey(dir, strings.ToLower(filename)), nil
}

// StaticKeyGenerator produces a fixed key, useful for singleton files
// like background music.
type StaticKeyGenerator struct {
	Category Category
	Name     string
}

func (g *StaticKeyGenerator) GenerateKey(_ Category, _ uint, _ string) (string, error) {
	name := strings.TrimSpace(g.Name)
	if name == "" {
		return "", fmt.Errorf("static key name is empty")
	}
	dir := CategoryDir(g.Category)
	return JoinKey(dir, name), nil
}

// ---------------------------------------------------------------------------
// URL resolution
// ---------------------------------------------------------------------------

// URLResolver maps a VireFS key to a publicly accessible URL.
// It is constructed once at startup by the DI layer based on the storage mode.
type URLResolver func(key string) string

// ---------------------------------------------------------------------------
// Path / key utilities
// ---------------------------------------------------------------------------

// JoinKey joins segments into a VireFS key (no leading slash).
func JoinKey(segments ...string) string {
	return strings.TrimPrefix(path.Join(segments...), "/")
}

// TrimLeadingSlash removes a leading "/" to convert a virtual path to a
// VireFS key.
func TrimLeadingSlash(p string) string {
	return strings.TrimPrefix(p, "/")
}
