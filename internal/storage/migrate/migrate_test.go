package migrate

import (
	"bytes"
	"context"
	"io"
	"testing"

	localStorage "github.com/lin-snow/ech0/internal/storage/local"
	stgx "github.com/lin-snow/ech0/pkg/storagex"
)

func TestCopy_LocalToLocal(t *testing.T) {
	src := localStorage.NewLocalFS(localStorage.WithRoot(t.TempDir()))
	dst := localStorage.NewLocalFS(localStorage.WithRoot(t.TempDir()))
	ctx := context.Background()

	_ = src.Write(ctx, "/images/a.png", bytes.NewReader([]byte("aaa")), stgx.WriteOptions{})
	_ = src.Write(ctx, "/images/b.png", bytes.NewReader([]byte("bbb")), stgx.WriteOptions{})

	result, err := Copy(ctx, src, dst, "/images", Options{Conflict: ConflictOverwrite})
	if err != nil {
		t.Fatalf("copy failed: %v", err)
	}
	if result.Copied != 2 {
		t.Fatalf("expected 2 copied, got %d", result.Copied)
	}

	rc, err := dst.Open(ctx, "/images/a.png")
	if err != nil {
		t.Fatalf("open failed: %v", err)
	}
	data, _ := io.ReadAll(rc)
	rc.Close()
	if string(data) != "aaa" {
		t.Fatalf("content mismatch: got %q", string(data))
	}
}

func TestCopy_SkipExisting(t *testing.T) {
	src := localStorage.NewLocalFS(localStorage.WithRoot(t.TempDir()))
	dst := localStorage.NewLocalFS(localStorage.WithRoot(t.TempDir()))
	ctx := context.Background()

	_ = src.Write(ctx, "/images/a.png", bytes.NewReader([]byte("src")), stgx.WriteOptions{})
	_ = dst.Write(ctx, "/images/a.png", bytes.NewReader([]byte("dst")), stgx.WriteOptions{})

	result, err := Copy(ctx, src, dst, "/images", Options{Conflict: ConflictSkip})
	if err != nil {
		t.Fatalf("copy failed: %v", err)
	}
	if result.Skipped != 1 {
		t.Fatalf("expected 1 skipped, got %d", result.Skipped)
	}

	rc, err := dst.Open(ctx, "/images/a.png")
	if err != nil {
		t.Fatalf("open failed: %v", err)
	}
	data, _ := io.ReadAll(rc)
	rc.Close()
	if string(data) != "dst" {
		t.Fatalf("expected original content preserved, got %q", string(data))
	}
}

func TestCopy_OverwriteExisting(t *testing.T) {
	src := localStorage.NewLocalFS(localStorage.WithRoot(t.TempDir()))
	dst := localStorage.NewLocalFS(localStorage.WithRoot(t.TempDir()))
	ctx := context.Background()

	_ = src.Write(ctx, "/images/a.png", bytes.NewReader([]byte("new")), stgx.WriteOptions{})
	_ = dst.Write(ctx, "/images/a.png", bytes.NewReader([]byte("old")), stgx.WriteOptions{})

	result, err := Copy(ctx, src, dst, "/images", Options{Conflict: ConflictOverwrite})
	if err != nil {
		t.Fatalf("copy failed: %v", err)
	}
	if result.Copied != 1 {
		t.Fatalf("expected 1 copied, got %d", result.Copied)
	}

	rc, err := dst.Open(ctx, "/images/a.png")
	if err != nil {
		t.Fatalf("open failed: %v", err)
	}
	data, _ := io.ReadAll(rc)
	rc.Close()
	if string(data) != "new" {
		t.Fatalf("expected overwritten content, got %q", string(data))
	}
}

func TestCopy_DryRun(t *testing.T) {
	src := localStorage.NewLocalFS(localStorage.WithRoot(t.TempDir()))
	dst := localStorage.NewLocalFS(localStorage.WithRoot(t.TempDir()))
	ctx := context.Background()

	_ = src.Write(ctx, "/images/a.png", bytes.NewReader([]byte("data")), stgx.WriteOptions{})

	result, err := Copy(ctx, src, dst, "/images", Options{DryRun: true})
	if err != nil {
		t.Fatalf("copy failed: %v", err)
	}
	if result.Copied != 1 {
		t.Fatalf("dry run should report 1 copied, got %d", result.Copied)
	}

	exists, _ := dst.Exists(ctx, "/images/a.png")
	if exists {
		t.Fatal("dry run should not actually copy files")
	}
}

func TestCopy_EmptySource(t *testing.T) {
	src := localStorage.NewLocalFS(localStorage.WithRoot(t.TempDir()))
	dst := localStorage.NewLocalFS(localStorage.WithRoot(t.TempDir()))

	result, err := Copy(context.Background(), src, dst, "/images", Options{})
	if err != nil {
		t.Fatalf("copy failed: %v", err)
	}
	if result.Copied != 0 {
		t.Fatalf("expected 0 copied, got %d", result.Copied)
	}
}
