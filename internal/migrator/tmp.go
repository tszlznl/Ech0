// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package migrator

import (
	"os"
	"path/filepath"
	"strings"
)

// TmpRelativeDir 是迁移上传源在 data/ 下的暂存目录(相对路径)。导入完成或清理时按此回收。
const TmpRelativeDir = "files/tmp"

// CleanupTmpDirFromPayload 按 source_payload.tmp_dir 安全删除暂存目录;无 tmp_dir 时幂等返回。
func CleanupTmpDirFromPayload(sourcePayload map[string]any) error {
	tmpDir, ok := resolveTmpDir(sourcePayload)
	if !ok {
		return nil
	}
	return os.RemoveAll(tmpDir)
}

// resolveTmpDir 把 source_payload.tmp_dir 解析为受限的绝对路径:必须落在 data/files/tmp 之下,
// 拒绝绝对路径 / 越级(..),防目录穿越。
func resolveTmpDir(sourcePayload map[string]any) (string, bool) {
	if len(sourcePayload) == 0 {
		return "", false
	}
	tmpDirRaw, ok := sourcePayload["tmp_dir"].(string)
	if !ok || strings.TrimSpace(tmpDirRaw) == "" {
		return "", false
	}
	cleanRelPath := filepath.Clean(filepath.FromSlash(strings.TrimSpace(tmpDirRaw)))
	if cleanRelPath == "." || cleanRelPath == "" || filepath.IsAbs(cleanRelPath) || strings.HasPrefix(cleanRelPath, "..") {
		return "", false
	}

	allowedBaseDir := filepath.Clean(filepath.Join("data", TmpRelativeDir))
	targetDir := filepath.Clean(filepath.Join("data", cleanRelPath))
	if targetDir != allowedBaseDir && !strings.HasPrefix(targetDir, allowedBaseDir+string(os.PathSeparator)) {
		return "", false
	}
	return targetDir, true
}
