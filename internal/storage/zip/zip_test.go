package zip

import (
	"testing"

	fileUtil "github.com/lin-snow/ech0/internal/util/file"
	"github.com/spf13/afero"
)

func TestModule_ZipUnzipCopy_WithMemMapFs(t *testing.T) {
	fs := afero.NewMemMapFs()
	module := NewModule(fs)

	if err := fs.MkdirAll("data/src/nested", 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := afero.WriteFile(fs, "data/src/nested/a.txt", []byte("hello"), 0o644); err != nil {
		t.Fatalf("write source file failed: %v", err)
	}

	if err := module.ZipDirectory("data/src", "data/archive.zip", fileUtil.ZipOptions{}); err != nil {
		t.Fatalf("zip failed: %v", err)
	}
	if err := module.UnzipFile("data/archive.zip", "data/unzipped"); err != nil {
		t.Fatalf("unzip failed: %v", err)
	}
	if err := module.CopyDirectory("data/unzipped/src", "data/copied"); err != nil {
		t.Fatalf("copy failed: %v", err)
	}

	content, err := afero.ReadFile(fs, "data/copied/nested/a.txt")
	if err != nil {
		t.Fatalf("read copied file failed: %v", err)
	}
	if string(content) != "hello" {
		t.Fatalf("unexpected copied content: %s", string(content))
	}
}

