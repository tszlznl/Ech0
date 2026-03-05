package local

import (
	"bytes"
	"context"
	"testing"

	"github.com/lin-snow/ech0/internal/storage"
	"github.com/spf13/afero"
)

type readSeekCloser struct {
	*bytes.Reader
}

func (r *readSeekCloser) Close() error { return nil }

func TestAdapter_SaveAndDelete_Image(t *testing.T) {
	fs := afero.NewMemMapFs()
	adapter := NewAdapterWithDirs(fs, "data/files/images", "data/files/audios")

	reader := &readSeekCloser{Reader: bytes.NewReader([]byte("fake-image-content"))}
	saved, err := adapter.Save(context.Background(), storage.SaveRequest{
		UserID:      1,
		FileName:    "test.png",
		ContentType: "image/png",
		Category:    storage.CategoryImage,
		Reader:      reader,
	})
	if err != nil {
		t.Fatalf("save image failed: %v", err)
	}
	if saved.Source != storage.SourceLocal {
		t.Fatalf("expected local source, got %s", saved.Source)
	}
	if saved.URL == "" {
		t.Fatalf("expected non-empty URL")
	}

	err = adapter.Delete(context.Background(), storage.DeleteRequest{
		URL:      saved.URL,
		Source:   storage.SourceLocal,
		Category: storage.CategoryImage,
	})
	if err != nil {
		t.Fatalf("delete image failed: %v", err)
	}
}

func TestAdapter_Save_AudioUsesMusicName(t *testing.T) {
	fs := afero.NewMemMapFs()
	adapter := NewAdapterWithDirs(fs, "data/files/images", "data/files/audios")

	rs := &readSeekCloser{Reader: bytes.NewReader([]byte("fake-audio-content"))}
	saved, err := adapter.Save(context.Background(), storage.SaveRequest{
		UserID:      1,
		FileName:    "any-name.mp3",
		ContentType: "audio/mpeg",
		Category:    storage.CategoryAudio,
		Reader:      rs,
	})
	if err != nil {
		t.Fatalf("save audio failed: %v", err)
	}
	if saved.URL != "/files/audios/music.mp3" {
		t.Fatalf("expected /files/audios/music.mp3, got %s", saved.URL)
	}
}

