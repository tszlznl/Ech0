package objectfs

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	stgx "github.com/lin-snow/ech0/pkg/storagex"
)

type fakeObjectStorage struct {
	uploaded map[string]string
}

func newFakeObjectStorage() *fakeObjectStorage {
	return &fakeObjectStorage{uploaded: make(map[string]string)}
}

func (f *fakeObjectStorage) Upload(_ context.Context, key string, r io.Reader, _ string) error {
	data, _ := io.ReadAll(r)
	f.uploaded[key] = string(data)
	return nil
}

func (f *fakeObjectStorage) Download(_ context.Context, key string) (io.ReadCloser, error) {
	content, ok := f.uploaded[key]
	if !ok {
		return nil, stgx.ErrNotFound
	}
	return io.NopCloser(strings.NewReader(content)), nil
}

func (f *fakeObjectStorage) ListObjects(_ context.Context, prefix string) ([]string, error) {
	var result []string
	for key := range f.uploaded {
		if strings.HasPrefix(key, prefix) {
			result = append(result, key)
		}
	}
	return result, nil
}

func (f *fakeObjectStorage) ListObjectStream(_ context.Context, _ string) (<-chan string, error) {
	ch := make(chan string)
	close(ch)
	return ch, nil
}

func (f *fakeObjectStorage) DeleteObject(_ context.Context, key string) error {
	delete(f.uploaded, key)
	return nil
}

func (f *fakeObjectStorage) PresignURL(_ context.Context, key string, _ time.Duration, method string) (string, error) {
	return "https://presigned/" + method + "/" + key, nil
}

func TestObjectFS_WriteAndOpen(t *testing.T) {
	fake := newFakeObjectStorage()
	fs := &ObjectFS{
		client:     fake,
		cfg:        stgx.ObjectStorageConfig{BucketName: "test-bucket", UseSSL: true, Endpoint: "s3.example.com"},
		pathPrefix: "uploads",
	}
	ctx := context.Background()

	err := fs.Write(ctx, "/images/test.png", strings.NewReader("hello"), stgx.WriteOptions{ContentType: "image/png"})
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}

	if _, ok := fake.uploaded["uploads/images/test.png"]; !ok {
		t.Fatal("expected key uploads/images/test.png in storage")
	}

	rc, err := fs.Open(ctx, "/images/test.png")
	if err != nil {
		t.Fatalf("open failed: %v", err)
	}
	defer rc.Close()
	data, _ := io.ReadAll(rc)
	if string(data) != "hello" {
		t.Fatalf("content mismatch: got %q", string(data))
	}
}

func TestObjectFS_Delete(t *testing.T) {
	fake := newFakeObjectStorage()
	fake.uploaded["uploads/images/test.png"] = "data"
	fs := &ObjectFS{client: fake, pathPrefix: "uploads"}
	ctx := context.Background()

	if err := fs.Delete(ctx, "/images/test.png"); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if _, ok := fake.uploaded["uploads/images/test.png"]; ok {
		t.Fatal("expected key to be deleted")
	}
}

func TestObjectFS_Exists(t *testing.T) {
	fake := newFakeObjectStorage()
	fake.uploaded["uploads/images/test.png"] = "data"
	fs := &ObjectFS{client: fake, pathPrefix: "uploads"}
	ctx := context.Background()

	ok, err := fs.Exists(ctx, "/images/test.png")
	if err != nil {
		t.Fatalf("exists failed: %v", err)
	}
	if !ok {
		t.Fatal("expected file to exist")
	}

	ok, err = fs.Exists(ctx, "/images/nope.png")
	if err != nil {
		t.Fatalf("exists failed: %v", err)
	}
	if ok {
		t.Fatal("expected file not to exist")
	}
}

func TestObjectFS_ResolveURL(t *testing.T) {
	fs := &ObjectFS{
		cfg: stgx.ObjectStorageConfig{
			BucketName: "my-bucket",
			Endpoint:   "s3.example.com",
			UseSSL:     true,
		},
		pathPrefix: "uploads",
	}
	url, err := fs.ResolveURL(context.Background(), "/images/test.png")
	if err != nil {
		t.Fatalf("resolve URL failed: %v", err)
	}
	want := "https://s3.example.com/my-bucket/uploads/images/test.png"
	if url != want {
		t.Fatalf("got %q, want %q", url, want)
	}
}

func TestObjectFS_ResolveURL_CDN(t *testing.T) {
	fs := &ObjectFS{
		cfg: stgx.ObjectStorageConfig{
			BucketName: "my-bucket",
			Endpoint:   "s3.example.com",
			UseSSL:     true,
			CDNURL:     "https://cdn.example.com",
		},
		pathPrefix: "uploads",
	}
	url, err := fs.ResolveURL(context.Background(), "/images/test.png")
	if err != nil {
		t.Fatalf("resolve URL failed: %v", err)
	}
	want := "https://cdn.example.com/uploads/images/test.png"
	if url != want {
		t.Fatalf("got %q, want %q", url, want)
	}
}

func TestObjectFS_ResolveURL_NoPrefix(t *testing.T) {
	fs := &ObjectFS{
		cfg: stgx.ObjectStorageConfig{
			BucketName: "my-bucket",
			Endpoint:   "s3.example.com",
			UseSSL:     true,
		},
	}
	url, err := fs.ResolveURL(context.Background(), "/images/test.png")
	if err != nil {
		t.Fatalf("resolve URL failed: %v", err)
	}
	want := "https://s3.example.com/my-bucket/images/test.png"
	if url != want {
		t.Fatalf("got %q, want %q", url, want)
	}
}

func TestObjectFS_Sign(t *testing.T) {
	fake := newFakeObjectStorage()
	fs := &ObjectFS{client: fake, pathPrefix: "uploads"}

	url, err := fs.Sign(context.Background(), "/images/test.png", "PUT", 24*time.Hour)
	if err != nil {
		t.Fatalf("sign failed: %v", err)
	}
	if url != "https://presigned/PUT/uploads/images/test.png" {
		t.Fatalf("unexpected presign URL: %s", url)
	}
}

func TestObjectFS_KeyToVirtualPath(t *testing.T) {
	fs := &ObjectFS{pathPrefix: "uploads"}
	got := fs.keyToVirtualPath("uploads/images/test.png")
	if got != "/images/test.png" {
		t.Fatalf("got %q, want /images/test.png", got)
	}
}

func TestObjectFS_KeyToVirtualPath_NoPrefix(t *testing.T) {
	fs := &ObjectFS{}
	got := fs.keyToVirtualPath("images/test.png")
	if got != "/images/test.png" {
		t.Fatalf("got %q, want /images/test.png", got)
	}
}
