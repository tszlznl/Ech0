// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package snapshot 是 Ech0 统一的快照资源:把整个 data/ 目录打成 zip 归档(导出端),
// 以及把外部上传的 zip 解开到目录(导入端)。它是 import / export 两端共用的同一产物,
// 只依赖 virefs 与配置,可被无 DI 的 CLI 直接调用。
//
// 快照格式即「data/ 的 zip」(排除 snapshots/、tmp/)。数据库文件名固定为 ech0.db
// (见 config 默认),导入端在 ech0 importer 中定位。
package snapshot

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"github.com/lin-snow/ech0/pkg/virefs"
	vizip "github.com/lin-snow/ech0/pkg/virefs/plugin/zip"
	"go.uber.org/zap"
)

// ErrNoSnapshot 表示快照目录下尚无可下载的快照（需先执行一次导出）。
var ErrNoSnapshot = errors.New("snapshot: no snapshot available")

const (
	dataDir             = "data"
	snapshotRelativeDir = "files/snapshots"
	tmpRelativeDir      = "files/tmp"
	snapshotFileName    = "ech0_snapshot"
	timeLayout          = "2006-01-02_15-04-05"
)

// Create packs the data/ directory into a zip snapshot using VireFS and returns
// the path and file name of the produced archive. It excludes the snapshots/ and
// tmp/ subtrees and keeps only the latest snapshot locally.
func Create() (string, string, error) {
	snapshotTime := time.Now().UTC().Format(timeLayout)
	fileName := fmt.Sprintf("%s_%s.zip", snapshotFileName, snapshotTime)
	snapshotDir := filepath.Join(dataDir, snapshotRelativeDir)
	snapshotPath := filepath.Join(snapshotDir, fileName)
	tempPath := filepath.Join(snapshotDir, "."+fileName+".tmp")

	if err := os.MkdirAll(snapshotDir, 0o755); err != nil {
		return "", "", fmt.Errorf("create snapshot dir: %w", err)
	}

	dataFS, err := virefs.NewLocalFS(dataDir)
	if err != nil {
		return "", "", fmt.Errorf("open data dir: %w", err)
	}

	ctx := context.Background()
	var keys []string
	if err := virefs.Walk(ctx, dataFS, "", func(key string, info virefs.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir {
			return nil
		}
		cleanKey := strings.Trim(strings.TrimSpace(key), "/")
		if cleanKey == "" {
			return nil
		}
		if shouldExcludeFromSnapshot(cleanKey) {
			return nil
		}
		keys = append(keys, cleanKey)
		return nil
	}); err != nil {
		return "", "", fmt.Errorf("walk data dir: %w", err)
	}

	f, err := os.Create(tempPath)
	if err != nil {
		return "", "", fmt.Errorf("create zip file: %w", err)
	}

	if err := vizip.Pack(ctx, dataFS, keys, f); err != nil {
		if closeErr := f.Close(); closeErr != nil {
			logUtil.GetLogger().Warn("Failed to close snapshot zip after pack error",
				zap.String("path", tempPath), zap.String("error", closeErr.Error()))
		}
		_ = os.Remove(tempPath)
		return "", "", fmt.Errorf("pack zip: %w", err)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tempPath)
		return "", "", fmt.Errorf("close zip file: %w", err)
	}

	if err := os.Rename(tempPath, snapshotPath); err != nil {
		_ = os.Remove(tempPath)
		return "", "", fmt.Errorf("finalize snapshot zip: %w", err)
	}

	if err := keepOnlyLatestSnapshot(snapshotDir, fileName); err != nil {
		return "", "", err
	}

	return snapshotPath, fileName, nil
}

// LatestPath 返回 data/files/snapshots 下最新一份快照 zip 的路径（文件名内嵌 UTC 时间戳，
// 取字典序最大者）。无可用快照时返回 ErrNoSnapshot。配合「仅保留最新一份」语义，目录里
// 通常只有一个 .zip。供同步下载出口取回「上一次导出作业产出的快照」。
func LatestPath() (string, error) {
	snapshotDir := filepath.Join(dataDir, snapshotRelativeDir)
	entries, err := os.ReadDir(snapshotDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrNoSnapshot
		}
		return "", fmt.Errorf("read snapshot dir: %w", err)
	}
	latest := ""
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// 跳过打包中的临时文件（"."+name+".tmp"），只认完成的 .zip。
		if !strings.HasSuffix(name, ".zip") {
			continue
		}
		if name > latest {
			latest = name
		}
	}
	if latest == "" {
		return "", ErrNoSnapshot
	}
	return filepath.Join(snapshotDir, latest), nil
}

// Unpack unpacks a snapshot zip file to the destination directory.
func Unpack(zipPath, destDir string) error {
	f, err := os.Open(zipPath)
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			logUtil.GetLogger().Warn("Failed to close snapshot zip reader",
				zap.String("path", zipPath), zap.String("error", closeErr.Error()))
		}
	}()

	info, err := f.Stat()
	if err != nil {
		return fmt.Errorf("stat zip: %w", err)
	}

	dstFS, err := virefs.NewLocalFS(destDir, virefs.WithCreateRoot())
	if err != nil {
		return fmt.Errorf("open dest dir: %w", err)
	}

	return vizip.Unpack(context.Background(), f, info.Size(), dstFS, "")
}

func shouldExcludeFromSnapshot(cleanKey string) bool {
	snapshotPrefix := strings.Trim(strings.TrimSpace(snapshotRelativeDir), "/")
	tmpPrefix := strings.Trim(strings.TrimSpace(tmpRelativeDir), "/")
	return cleanKey == snapshotPrefix ||
		strings.HasPrefix(cleanKey, snapshotPrefix+"/") ||
		cleanKey == tmpPrefix ||
		strings.HasPrefix(cleanKey, tmpPrefix+"/")
}

func keepOnlyLatestSnapshot(snapshotDir string, latestFileName string) error {
	entries, err := os.ReadDir(snapshotDir)
	if err != nil {
		return fmt.Errorf("read snapshot dir: %w", err)
	}
	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})
	for _, entry := range entries {
		name := entry.Name()
		if name == latestFileName {
			continue
		}
		removePath := filepath.Join(snapshotDir, name)
		if err := os.RemoveAll(removePath); err != nil {
			return fmt.Errorf("cleanup old snapshot %s: %w", name, err)
		}
	}
	return nil
}
