package backup

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecuteBackup_KeepOnlyLatestAndExcludeBackupDir(t *testing.T) {
	workspace := t.TempDir()
	prevWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(prevWD)
	})
	if err := os.Chdir(workspace); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(dataDir, "files/images"), 0o755); err != nil {
		t.Fatalf("mkdir images failed: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dataDir, backupRelativeDir), 0o755); err != nil {
		t.Fatalf("mkdir backup dir failed: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dataDir, "ech0.db"), []byte("db"), 0o644); err != nil {
		t.Fatalf("write ech0.db failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "files/images/demo.txt"), []byte("demo"), 0o644); err != nil {
		t.Fatalf("write image file failed: %v", err)
	}
	legacyBackupPath := filepath.Join(dataDir, backupRelativeDir, "legacy.zip")
	if err := os.WriteFile(legacyBackupPath, []byte("legacy"), 0o644); err != nil {
		t.Fatalf("write legacy backup failed: %v", err)
	}

	backupPath, fileName, err := ExecuteBackup()
	if err != nil {
		t.Fatalf("ExecuteBackup failed: %v", err)
	}
	if !strings.HasPrefix(backupPath, filepath.Join(dataDir, backupRelativeDir)) {
		t.Fatalf("backup path should be under %s, got %s", filepath.Join(dataDir, backupRelativeDir), backupPath)
	}
	if fileName == "" {
		t.Fatal("backup filename should not be empty")
	}

	entries, err := os.ReadDir(filepath.Join(dataDir, backupRelativeDir))
	if err != nil {
		t.Fatalf("readdir backup dir failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("backup dir should keep only one file, got %d", len(entries))
	}
	if entries[0].Name() != fileName {
		t.Fatalf("kept file should be latest backup %s, got %s", fileName, entries[0].Name())
	}

	zr, err := zip.OpenReader(backupPath)
	if err != nil {
		t.Fatalf("open backup zip failed: %v", err)
	}
	defer func() { _ = zr.Close() }()

	for _, f := range zr.File {
		if strings.HasPrefix(f.Name, backupRelativeDir+"/") || f.Name == backupRelativeDir {
			t.Fatalf("backup zip should exclude %s, got entry: %s", backupRelativeDir, f.Name)
		}
	}
}

func TestShouldExcludeFromBackup(t *testing.T) {
	tests := []struct {
		key      string
		expected bool
	}{
		{key: "files/backups", expected: true},
		{key: "files/backups/a.zip", expected: true},
		{key: "files/images/a.png", expected: false},
		{key: "ech0.db", expected: false},
	}

	for _, tc := range tests {
		got := shouldExcludeFromBackup(tc.key)
		if got != tc.expected {
			t.Fatalf("shouldExcludeFromBackup(%q)=%v, expected %v", tc.key, got, tc.expected)
		}
	}
}
