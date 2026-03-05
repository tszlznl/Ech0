package backup

import (
	"testing"

	"github.com/spf13/afero"
)

func TestExecuteBackupWithMemMapFs(t *testing.T) {
	fs := afero.NewMemMapFs()
	if err := fs.MkdirAll("data", 0o755); err != nil {
		t.Fatalf("mkdir data failed: %v", err)
	}
	if err := afero.WriteFile(fs, "data/ech0.db", []byte("db"), 0o644); err != nil {
		t.Fatalf("write db file failed: %v", err)
	}

	backupPath, backupName, err := ExecuteBackup(fs)
	if err != nil {
		t.Fatalf("execute backup failed: %v", err)
	}
	if backupPath == "" || backupName == "" {
		t.Fatal("backup path/name should not be empty")
	}
	if _, err := fs.Stat(backupPath); err != nil {
		t.Fatalf("backup zip should exist: %v", err)
	}
}

