package local

import (
	"bytes"
	"context"
	"io"
	"testing"

	stgx "github.com/lin-snow/ech0/pkg/storagex"
)

func TestLocalFS_WriteAndOpen(t *testing.T) {
	fs := NewLocalFS(WithRoot(t.TempDir()))
	ctx := context.Background()

	content := []byte("hello world")
	err := fs.Write(ctx, "/images/test.png", bytes.NewReader(content), stgx.WriteOptions{ContentType: "image/png"})
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}

	rc, err := fs.Open(ctx, "/images/test.png")
	if err != nil {
		t.Fatalf("open failed: %v", err)
	}
	defer rc.Close()

	buf, _ := io.ReadAll(rc)
	if string(buf) != "hello world" {
		t.Fatalf("content mismatch: got %q", string(buf))
	}
}

func TestLocalFS_Delete(t *testing.T) {
	fs := NewLocalFS(WithRoot(t.TempDir()))
	ctx := context.Background()

	_ = fs.Write(ctx, "/images/test.png", bytes.NewReader([]byte("data")), stgx.WriteOptions{})

	exists, _ := fs.Exists(ctx, "/images/test.png")
	if !exists {
		t.Fatal("file should exist")
	}

	if err := fs.Delete(ctx, "/images/test.png"); err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	exists, _ = fs.Exists(ctx, "/images/test.png")
	if exists {
		t.Fatal("file should not exist after delete")
	}
}

func TestLocalFS_DeleteNonexistent(t *testing.T) {
	fs := NewLocalFS(WithRoot(t.TempDir()))
	if err := fs.Delete(context.Background(), "/nope.txt"); err != nil {
		t.Fatalf("deleting nonexistent file should not error: %v", err)
	}
}

func TestLocalFS_Stat(t *testing.T) {
	fs := NewLocalFS(WithRoot(t.TempDir()))
	ctx := context.Background()

	_ = fs.Write(ctx, "/images/test.png", bytes.NewReader([]byte("data")), stgx.WriteOptions{})

	info, err := fs.Stat(ctx, "/images/test.png")
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	if info.Size != 4 {
		t.Fatalf("unexpected size: %d", info.Size)
	}
	if info.Path != "/images/test.png" {
		t.Fatalf("unexpected path: %s", info.Path)
	}
}

func TestLocalFS_StatNotFound(t *testing.T) {
	fs := NewLocalFS(WithRoot(t.TempDir()))
	_, err := fs.Stat(context.Background(), "/nope.txt")
	if err != stgx.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestLocalFS_List(t *testing.T) {
	fs := NewLocalFS(WithRoot(t.TempDir()))
	ctx := context.Background()

	_ = fs.Write(ctx, "/images/a.png", bytes.NewReader([]byte("a")), stgx.WriteOptions{})
	_ = fs.Write(ctx, "/images/b.png", bytes.NewReader([]byte("b")), stgx.WriteOptions{})

	infos, err := fs.List(ctx, "/images")
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(infos) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(infos))
	}
}

func TestLocalFS_ListEmpty(t *testing.T) {
	fs := NewLocalFS(WithRoot(t.TempDir()))
	infos, err := fs.List(context.Background(), "/nonexistent")
	if err != nil {
		t.Fatalf("list nonexistent should return nil, got error: %v", err)
	}
	if infos != nil {
		t.Fatalf("expected nil, got %v", infos)
	}
}

func TestLocalFS_ResolveURL(t *testing.T) {
	fs := NewLocalFS(WithRoot(t.TempDir()))
	url, err := fs.ResolveURL(context.Background(), "/images/test.png")
	if err != nil {
		t.Fatalf("resolve URL failed: %v", err)
	}
	if url != "/files/images/test.png" {
		t.Fatalf("unexpected URL: %s", url)
	}
}

func TestLocalFS_Exists(t *testing.T) {
	fs := NewLocalFS(WithRoot(t.TempDir()))
	ctx := context.Background()

	ok, _ := fs.Exists(ctx, "/nope.txt")
	if ok {
		t.Fatal("should not exist")
	}

	_ = fs.Write(ctx, "/images/a.png", bytes.NewReader([]byte("a")), stgx.WriteOptions{})
	ok, _ = fs.Exists(ctx, "/images/a.png")
	if !ok {
		t.Fatal("should exist")
	}
}

func TestLocalFS_InvalidPath(t *testing.T) {
	fs := NewLocalFS(WithRoot(t.TempDir()))
	ctx := context.Background()

	if err := fs.Write(ctx, "/images/../../../etc/passwd", bytes.NewReader(nil), stgx.WriteOptions{}); err == nil {
		t.Fatal("expected error for traversal path")
	}
}
