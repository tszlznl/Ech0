package util

import (
	"testing"

	"github.com/spf13/afero"
)

func TestZipAndUnzipWithMemMapFs(t *testing.T) {
	fs := afero.NewMemMapFs()
	if err := fs.MkdirAll("data/nested", 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := afero.WriteFile(fs, "data/nested/a.txt", []byte("hello"), 0o644); err != nil {
		t.Fatalf("write file failed: %v", err)
	}

	if err := ZipDirectoryWithOptions(fs, "data", "backup/test.zip", DefaultZipOptions()); err != nil {
		t.Fatalf("zip failed: %v", err)
	}

	if err := UnzipFile(fs, "backup/test.zip", "restore"); err != nil {
		t.Fatalf("unzip failed: %v", err)
	}

	content, err := afero.ReadFile(fs, "restore/nested/a.txt")
	if err != nil {
		t.Fatalf("read restored file failed: %v", err)
	}
	if string(content) != "hello" {
		t.Fatalf("unexpected restored content: %s", string(content))
	}
}

func TestCopyDirectoryWithMemMapFs(t *testing.T) {
	fs := afero.NewMemMapFs()
	if err := fs.MkdirAll("src/x", 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := afero.WriteFile(fs, "src/x/file.md", []byte("# title"), 0o644); err != nil {
		t.Fatalf("write file failed: %v", err)
	}

	if err := CopyDirectory(fs, "src", "dst"); err != nil {
		t.Fatalf("copy directory failed: %v", err)
	}

	got, err := afero.ReadFile(fs, "dst/x/file.md")
	if err != nil {
		t.Fatalf("read copied file failed: %v", err)
	}
	if string(got) != "# title" {
		t.Fatalf("unexpected copied content: %s", string(got))
	}
}

func TestValidateAndSanitizePath(t *testing.T) {
	p, err := ValidateAndSanitizePath("data/files/images", "/files/images/demo.png", "/files/images/")
	if err != nil {
		t.Fatalf("sanitize should succeed: %v", err)
	}
	if p == "" {
		t.Fatal("sanitized path should not be empty")
	}

	if _, err := ValidateAndSanitizePath("data/files/images", "/files/images/../../passwd", "/files/images/"); err == nil {
		t.Fatal("path traversal should be rejected")
	}
}

