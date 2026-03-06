package storagex

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"
)

// KeyGenerator produces a virtual path for a new file upload
// based on category, user, and original filename.
type KeyGenerator interface {
	GenerateKey(category Category, userID uint, originalFilename string) (string, error)
}

// RandomKeyGenerator creates paths like /images/1_1700000000_abc123.png
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
	return JoinPath(dir, strings.ToLower(filename)), nil
}

// StaticKeyGenerator produces a fixed path, useful for singleton files
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
	return JoinPath(dir, name), nil
}
