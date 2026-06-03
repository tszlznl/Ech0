// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package snapshot

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreate_KeepOnlyLatestAndExcludeSnapshotDir(t *testing.T) {
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
	if err := os.MkdirAll(filepath.Join(dataDir, snapshotRelativeDir), 0o755); err != nil {
		t.Fatalf("mkdir snapshot dir failed: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dataDir, "ech0.db"), []byte("db"), 0o644); err != nil {
		t.Fatalf("write ech0.db failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "files/images/demo.txt"), []byte("demo"), 0o644); err != nil {
		t.Fatalf("write image file failed: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dataDir, tmpRelativeDir), 0o755); err != nil {
		t.Fatalf("mkdir tmp dir failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, tmpRelativeDir, "payload.txt"), []byte("tmp payload"), 0o644); err != nil {
		t.Fatalf("write tmp file failed: %v", err)
	}
	legacySnapshotPath := filepath.Join(dataDir, snapshotRelativeDir, "legacy.zip")
	if err := os.WriteFile(legacySnapshotPath, []byte("legacy"), 0o644); err != nil {
		t.Fatalf("write legacy snapshot failed: %v", err)
	}

	snapshotPath, fileName, err := Create()
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if !strings.HasPrefix(snapshotPath, filepath.Join(dataDir, snapshotRelativeDir)) {
		t.Fatalf("snapshot path should be under %s, got %s", filepath.Join(dataDir, snapshotRelativeDir), snapshotPath)
	}
	if fileName == "" {
		t.Fatal("snapshot filename should not be empty")
	}

	entries, err := os.ReadDir(filepath.Join(dataDir, snapshotRelativeDir))
	if err != nil {
		t.Fatalf("readdir snapshot dir failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("snapshot dir should keep only one file, got %d", len(entries))
	}
	if entries[0].Name() != fileName {
		t.Fatalf("kept file should be latest snapshot %s, got %s", fileName, entries[0].Name())
	}

	zr, err := zip.OpenReader(snapshotPath)
	if err != nil {
		t.Fatalf("open snapshot zip failed: %v", err)
	}
	defer func() { _ = zr.Close() }()

	for _, f := range zr.File {
		if strings.HasPrefix(f.Name, snapshotRelativeDir+"/") || f.Name == snapshotRelativeDir {
			t.Fatalf("snapshot zip should exclude %s, got entry: %s", snapshotRelativeDir, f.Name)
		}
		if strings.HasPrefix(f.Name, tmpRelativeDir+"/") || f.Name == tmpRelativeDir {
			t.Fatalf("snapshot zip should exclude %s, got entry: %s", tmpRelativeDir, f.Name)
		}
	}
}

func TestShouldExcludeFromSnapshot(t *testing.T) {
	tests := []struct {
		key      string
		expected bool
	}{
		{key: "files/snapshots", expected: true},
		{key: "files/snapshots/a.zip", expected: true},
		{key: "files/tmp", expected: true},
		{key: "files/tmp/a.zip", expected: true},
		{key: "files/images/a.png", expected: false},
		{key: "ech0.db", expected: false},
	}

	for _, tc := range tests {
		got := shouldExcludeFromSnapshot(tc.key)
		if got != tc.expected {
			t.Fatalf("shouldExcludeFromSnapshot(%q)=%v, expected %v", tc.key, got, tc.expected)
		}
	}
}
