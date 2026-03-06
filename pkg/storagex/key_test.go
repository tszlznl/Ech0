package storagex

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestRandomKeyGenerator(t *testing.T) {
	fixed := bytes.NewReader([]byte{0xAA, 0xBB, 0xCC, 0xDD})
	g := &RandomKeyGenerator{
		RandSource:  fixed,
		SuffixBytes: 4,
		Now:         func() time.Time { return time.Unix(1700000000, 0) },
	}
	got, err := g.GenerateKey(CategoryImage, 1, "PHOTO.JPG")
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}
	want := "/images/1_1700000000_aabbccdd.jpg"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestRandomKeyGeneratorKeepsExtension(t *testing.T) {
	g := NewRandomKeyGenerator()
	got, err := g.GenerateKey(CategoryAudio, 2, "SONG.MP3")
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}
	if !strings.HasPrefix(got, "/audios/") {
		t.Fatalf("expected /audios/ prefix, got %s", got)
	}
	if !strings.HasSuffix(got, ".mp3") {
		t.Fatalf("expected .mp3 suffix, got %s", got)
	}
}

func TestRandomKeyGeneratorNoExtension(t *testing.T) {
	g := NewRandomKeyGenerator()
	got, err := g.GenerateKey(CategoryFile, 1, "noext")
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}
	if !strings.HasSuffix(got, ".bin") {
		t.Fatalf("expected .bin suffix for extensionless file, got %s", got)
	}
}

func TestStaticKeyGenerator(t *testing.T) {
	g := &StaticKeyGenerator{Category: CategoryAudio, Name: "music.mp3"}
	got, err := g.GenerateKey(CategoryAudio, 0, "any.mp3")
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}
	if got != "/audios/music.mp3" {
		t.Fatalf("got %q, want /audios/music.mp3", got)
	}
}

func TestStaticKeyGeneratorEmpty(t *testing.T) {
	g := &StaticKeyGenerator{Category: CategoryAudio, Name: ""}
	_, err := g.GenerateKey(CategoryAudio, 0, "any.mp3")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestCategoryDir(t *testing.T) {
	tests := []struct {
		cat  Category
		want string
	}{
		{CategoryImage, "images"},
		{CategoryAudio, "audios"},
		{CategoryVideo, "videos"},
		{CategoryPDF, "documents"},
		{CategoryMarkdown, "documents"},
		{CategoryFile, "files"},
	}
	for _, tt := range tests {
		if got := CategoryDir(tt.cat); got != tt.want {
			t.Fatalf("CategoryDir(%s) = %q, want %q", tt.cat, got, tt.want)
		}
	}
}
