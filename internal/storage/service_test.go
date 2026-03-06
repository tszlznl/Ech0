package storage

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	stgx "github.com/lin-snow/ech0/pkg/storagex"
)

// memFS is a minimal in-memory FS for testing without disk I/O.
type memFS struct {
	files map[string][]byte
}

func newMemFS() *memFS {
	return &memFS{files: make(map[string][]byte)}
}

func (m *memFS) Open(_ context.Context, path string) (io.ReadCloser, error) {
	p, err := stgx.NormalizePath(path)
	if err != nil {
		return nil, err
	}
	data, ok := m.files[p]
	if !ok {
		return nil, stgx.ErrNotFound
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (m *memFS) Write(_ context.Context, path string, r io.Reader, _ stgx.WriteOptions) error {
	p, err := stgx.NormalizePath(path)
	if err != nil {
		return err
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	m.files[p] = data
	return nil
}

func (m *memFS) Delete(_ context.Context, path string) error {
	p, err := stgx.NormalizePath(path)
	if err != nil {
		return err
	}
	delete(m.files, p)
	return nil
}

func (m *memFS) Stat(_ context.Context, path string) (*stgx.FileInfo, error) {
	p, err := stgx.NormalizePath(path)
	if err != nil {
		return nil, err
	}
	data, ok := m.files[p]
	if !ok {
		return nil, stgx.ErrNotFound
	}
	return &stgx.FileInfo{Path: p, Size: int64(len(data))}, nil
}

func (m *memFS) List(_ context.Context, prefix string) ([]stgx.FileInfo, error) {
	p, _ := stgx.NormalizePath(prefix)
	var result []stgx.FileInfo
	for k, v := range m.files {
		if strings.HasPrefix(k, p+"/") || k == p {
			result = append(result, stgx.FileInfo{Path: k, Size: int64(len(v))})
		}
	}
	return result, nil
}

func (m *memFS) Exists(_ context.Context, path string) (bool, error) {
	p, err := stgx.NormalizePath(path)
	if err != nil {
		return false, nil
	}
	_, ok := m.files[p]
	return ok, nil
}

func (m *memFS) ResolveURL(_ context.Context, path string) (string, error) {
	p, err := stgx.NormalizePath(path)
	if err != nil {
		return "", err
	}
	return "/files" + p, nil
}

// signerFS wraps memFS and adds Sign support.
type signerFS struct {
	*memFS
}

func (s *signerFS) Sign(_ context.Context, path string, method string, _ time.Duration) (string, error) {
	return "https://presigned/" + method + "/" + stgx.TrimVirtualPath(path), nil
}

func TestStorageService_UploadAndDelete(t *testing.T) {
	fs := newMemFS()
	svc := NewStorageService(StorageServiceConfig{
		FS:     fs,
		Source: "local",
	})
	ctx := context.Background()

	result, err := svc.Upload(ctx, stgx.CategoryImage, 1, "photo.png", "image/png", bytes.NewReader([]byte("image-data")))
	if err != nil {
		t.Fatalf("upload failed: %v", err)
	}
	if !strings.HasPrefix(result.URL, "/files/images/") {
		t.Fatalf("expected URL starting with /files/images/, got %s", result.URL)
	}
	if result.ObjectKey == "" {
		t.Fatal("expected non-empty object key")
	}

	exists, _ := svc.Exists(ctx, result.VirtualPath)
	if !exists {
		t.Fatal("file should exist after upload")
	}

	if err := svc.Delete(ctx, result.VirtualPath); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	exists, _ = svc.Exists(ctx, result.VirtualPath)
	if exists {
		t.Fatal("file should not exist after delete")
	}
}

func TestStorageService_UploadAudioUsesStaticKey(t *testing.T) {
	fs := newMemFS()
	svc := NewStorageService(StorageServiceConfig{FS: fs, Source: "local"})

	result, err := svc.Upload(context.Background(), stgx.CategoryAudio, 1, "song.mp3", "audio/mpeg", bytes.NewReader([]byte("audio")))
	if err != nil {
		t.Fatalf("upload failed: %v", err)
	}
	if result.URL != "/files/audios/music.mp3" {
		t.Fatalf("expected /files/audios/music.mp3, got %s", result.URL)
	}
}

func TestStorageService_ResolveURL(t *testing.T) {
	fs := newMemFS()
	svc := NewStorageService(StorageServiceConfig{FS: fs, Source: "local"})

	url, err := svc.ResolveURL(context.Background(), "/images/test.png")
	if err != nil {
		t.Fatalf("resolve URL failed: %v", err)
	}
	if url != "/files/images/test.png" {
		t.Fatalf("expected /files/images/test.png, got %s", url)
	}
}

func TestStorageService_Presign(t *testing.T) {
	fs := &signerFS{newMemFS()}
	svc := NewStorageService(StorageServiceConfig{FS: fs, Source: "s3"})

	result, err := svc.Presign(context.Background(), stgx.CategoryImage, 1, "photo.png", "image/png", "PUT", 24*time.Hour)
	if err != nil {
		t.Fatalf("presign failed: %v", err)
	}
	if result.PresignURL == "" {
		t.Fatal("expected non-empty presign URL")
	}
	if result.FileURL == "" {
		t.Fatal("expected non-empty file URL")
	}
	if result.ObjectKey == "" {
		t.Fatal("expected non-empty object key")
	}
}

func TestStorageService_PresignUnsupported(t *testing.T) {
	fs := newMemFS()
	svc := NewStorageService(StorageServiceConfig{FS: fs, Source: "local"})

	_, err := svc.Presign(context.Background(), stgx.CategoryImage, 1, "photo.png", "image/png", "PUT", 0)
	if err == nil {
		t.Fatal("expected error for presign on non-signer backend")
	}
}

func TestStorageService_Source(t *testing.T) {
	svc := NewStorageService(StorageServiceConfig{FS: newMemFS(), Source: "s3"})
	if svc.Source() != "s3" {
		t.Fatalf("expected s3, got %s", svc.Source())
	}
}

func TestStorageService_OpenAndStat(t *testing.T) {
	fs := newMemFS()
	svc := NewStorageService(StorageServiceConfig{FS: fs, Source: "local"})
	ctx := context.Background()

	result, _ := svc.Upload(ctx, stgx.CategoryImage, 1, "photo.png", "image/png", bytes.NewReader([]byte("hello")))

	rc, err := svc.Open(ctx, result.VirtualPath)
	if err != nil {
		t.Fatalf("open failed: %v", err)
	}
	data, _ := io.ReadAll(rc)
	rc.Close()
	if string(data) != "hello" {
		t.Fatalf("content mismatch: got %q", string(data))
	}

	info, err := svc.Stat(ctx, result.VirtualPath)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	if info.Size != 5 {
		t.Fatalf("expected size 5, got %d", info.Size)
	}
}
