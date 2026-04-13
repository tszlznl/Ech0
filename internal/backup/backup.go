package backup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	virefs "github.com/lin-snow/VireFS"
	vizip "github.com/lin-snow/VireFS/plugin/zip"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"go.uber.org/zap"
)

const (
	dataDir           = "data"
	backupRelativeDir = "files/backups"
	tmpRelativeDir    = "files/tmp"
	backupFileName    = "ech0_backup"
	timeLayout        = "2006-01-02_15-04-05"
)

// ExecuteBackup packs the data/ directory into a zip archive using VireFS.
func ExecuteBackup() (string, string, error) {
	backupTime := time.Now().UTC().Format(timeLayout)
	fileName := fmt.Sprintf("%s_%s.zip", backupFileName, backupTime)
	backupDir := filepath.Join(dataDir, backupRelativeDir)
	backupPath := filepath.Join(backupDir, fileName)
	tempPath := filepath.Join(backupDir, "."+fileName+".tmp")

	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return "", "", fmt.Errorf("create backup dir: %w", err)
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
		if shouldExcludeFromBackup(cleanKey) {
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
			logUtil.GetLogger().Warn("Failed to close backup zip after pack error",
				zap.String("path", tempPath), zap.String("error", closeErr.Error()))
		}
		_ = os.Remove(tempPath)
		return "", "", fmt.Errorf("pack zip: %w", err)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tempPath)
		return "", "", fmt.Errorf("close zip file: %w", err)
	}

	if err := os.Rename(tempPath, backupPath); err != nil {
		_ = os.Remove(tempPath)
		return "", "", fmt.Errorf("finalize backup zip: %w", err)
	}

	if err := keepOnlyLatestBackup(backupDir, fileName); err != nil {
		return "", "", err
	}

	return backupPath, fileName, nil
}

// UnpackZipToDir unpacks a zip file to destination directory.
func UnpackZipToDir(zipPath, destDir string) error {
	f, err := os.Open(zipPath)
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			logUtil.GetLogger().Warn("Failed to close backup zip reader",
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

func shouldExcludeFromBackup(cleanKey string) bool {
	backupPrefix := strings.Trim(strings.TrimSpace(backupRelativeDir), "/")
	tmpPrefix := strings.Trim(strings.TrimSpace(tmpRelativeDir), "/")
	return cleanKey == backupPrefix ||
		strings.HasPrefix(cleanKey, backupPrefix+"/") ||
		cleanKey == tmpPrefix ||
		strings.HasPrefix(cleanKey, tmpPrefix+"/")
}

func keepOnlyLatestBackup(backupDir string, latestFileName string) error {
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return fmt.Errorf("read backup dir: %w", err)
	}
	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})
	for _, entry := range entries {
		name := entry.Name()
		if name == latestFileName {
			continue
		}
		removePath := filepath.Join(backupDir, name)
		if err := os.RemoveAll(removePath); err != nil {
			return fmt.Errorf("cleanup old backup %s: %w", name, err)
		}
	}
	return nil
}
