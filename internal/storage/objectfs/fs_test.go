package objectfs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/lin-snow/ech0/pkg/s3x"
	stgx "github.com/lin-snow/ech0/pkg/storagex"
)

// fakeClient implements s3x.Client for testing.
type fakeClient struct {
	objects map[string][]byte
}

func newFakeClient() *fakeClient {
	return &fakeClient{objects: make(map[string][]byte)}
}

func (f *fakeClient) PutObject(_ context.Context, _, key string, body io.Reader, _ string) error {
	data, _ := io.ReadAll(body)
	f.objects[key] = data
	return nil
}

func (f *fakeClient) GetObject(_ context.Context, _, key string) (io.ReadCloser, error) {
	data, ok := f.objects[key]
	if !ok {
		return nil, fmt.Errorf("NoSuchKey")
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (f *fakeClient) DeleteObject(_ context.Context, _, key string) error {
	delete(f.objects, key)
	return nil
}

func (f *fakeClient) HeadObject(_ context.Context, _, key string) (*s3x.ObjectInfo, error) {
	data, ok := f.objects[key]
	if !ok {
		return nil, fmt.Errorf("NotFound")
	}
	return &s3x.ObjectInfo{Key: key, Size: int64(len(data))}, nil
}

func (f *fakeClient) ListObjects(_ context.Context, _, prefix string) ([]s3x.ObjectEntry, error) {
	var entries []s3x.ObjectEntry
	for key, data := range f.objects {
		if strings.HasPrefix(key, prefix) {
			entries = append(entries, s3x.ObjectEntry{Key: key, Size: int64(len(data))})
		}
	}
	return entries, nil
}

func (f *fakeClient) PresignGetObject(_ context.Context, _, key string, _ time.Duration) (string, error) {
	return "https://presigned/GET/" + key, nil
}

func (f *fakeClient) PresignPutObject(_ context.Context, _, key string, _ time.Duration) (string, error) {
	return "https://presigned/PUT/" + key, nil
}

func newTestFS(fake *fakeClient, pathPrefix string, cfg stgx.ObjectStorageConfig) *ObjectFS {
	return New(fake, cfg, WithPathPrefix(pathPrefix))
}

func TestObjectFS_WriteAndOpen(t *testing.T) {
	fake := newFakeClient()
	fs := newTestFS(fake, "uploads", stgx.ObjectStorageConfig{
		BucketName: "test-bucket",
		UseSSL:     true,
		Endpoint:   "s3.example.com",
	})
	ctx := context.Background()

	err := fs.Write(ctx, "/images/test.png", strings.NewReader("hello"), stgx.WriteOptions{ContentType: "image/png"})
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}

	if _, ok := fake.objects["uploads/images/test.png"]; !ok {
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
	fake := newFakeClient()
	fake.objects["uploads/images/test.png"] = []byte("data")
	fs := newTestFS(fake, "uploads", stgx.ObjectStorageConfig{BucketName: "b"})

	if err := fs.Delete(context.Background(), "/images/test.png"); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if _, ok := fake.objects["uploads/images/test.png"]; ok {
		t.Fatal("expected key to be deleted")
	}
}

func TestObjectFS_Exists(t *testing.T) {
	fake := newFakeClient()
	fake.objects["uploads/images/test.png"] = []byte("data")
	fs := newTestFS(fake, "uploads", stgx.ObjectStorageConfig{BucketName: "b"})
	ctx := context.Background()

	ok, _ := fs.Exists(ctx, "/images/test.png")
	if !ok {
		t.Fatal("expected file to exist")
	}

	ok, _ = fs.Exists(ctx, "/images/nope.png")
	if ok {
		t.Fatal("expected file not to exist")
	}
}

func TestObjectFS_Stat(t *testing.T) {
	fake := newFakeClient()
	fake.objects["uploads/images/test.png"] = []byte("hello")
	fs := newTestFS(fake, "uploads", stgx.ObjectStorageConfig{BucketName: "b"})

	info, err := fs.Stat(context.Background(), "/images/test.png")
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	if info.Size != 5 {
		t.Fatalf("expected size 5, got %d", info.Size)
	}
	if info.Path != "/images/test.png" {
		t.Fatalf("expected path /images/test.png, got %s", info.Path)
	}
}

func TestObjectFS_StatNotFound(t *testing.T) {
	fs := newTestFS(newFakeClient(), "uploads", stgx.ObjectStorageConfig{BucketName: "b"})
	_, err := fs.Stat(context.Background(), "/nope.txt")
	if err != stgx.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestObjectFS_List(t *testing.T) {
	fake := newFakeClient()
	fake.objects["uploads/images/a.png"] = []byte("a")
	fake.objects["uploads/images/b.png"] = []byte("b")
	fs := newTestFS(fake, "uploads", stgx.ObjectStorageConfig{BucketName: "b"})

	infos, err := fs.List(context.Background(), "/images")
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(infos) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(infos))
	}
}

func TestObjectFS_ResolveURL(t *testing.T) {
	fs := newTestFS(newFakeClient(), "uploads", stgx.ObjectStorageConfig{
		BucketName: "my-bucket",
		Endpoint:   "s3.example.com",
		UseSSL:     true,
	})
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
	fs := newTestFS(newFakeClient(), "uploads", stgx.ObjectStorageConfig{
		BucketName: "my-bucket",
		Endpoint:   "s3.example.com",
		UseSSL:     true,
		CDNURL:     "https://cdn.example.com",
	})
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
	fs := newTestFS(newFakeClient(), "", stgx.ObjectStorageConfig{
		BucketName: "my-bucket",
		Endpoint:   "s3.example.com",
		UseSSL:     true,
	})
	url, err := fs.ResolveURL(context.Background(), "/images/test.png")
	if err != nil {
		t.Fatalf("resolve URL failed: %v", err)
	}
	want := "https://s3.example.com/my-bucket/images/test.png"
	if url != want {
		t.Fatalf("got %q, want %q", url, want)
	}
}

func TestObjectFS_Sign_GET(t *testing.T) {
	fs := newTestFS(newFakeClient(), "uploads", stgx.ObjectStorageConfig{BucketName: "b"})
	url, err := fs.Sign(context.Background(), "/images/test.png", "GET", 24*time.Hour)
	if err != nil {
		t.Fatalf("sign failed: %v", err)
	}
	if url != "https://presigned/GET/uploads/images/test.png" {
		t.Fatalf("unexpected presign URL: %s", url)
	}
}

func TestObjectFS_Sign_PUT(t *testing.T) {
	fs := newTestFS(newFakeClient(), "uploads", stgx.ObjectStorageConfig{BucketName: "b"})
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
