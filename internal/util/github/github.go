// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package util

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/v80/github"
	"golang.org/x/mod/semver"
)

// isHelmChartArtifactStyleTag reports tags used only for Helm chart packages (ech0-<anything>),
// which must not be treated as application semver releases.
func isHelmChartArtifactStyleTag(tag string) bool {
	t := strings.TrimSpace(tag)
	return len(t) >= 5 && strings.EqualFold(t[:5], "ech0-")
}

func canonicalStableSemverFromReleaseTag(tag string) string {
	t := strings.TrimSpace(tag)
	if t == "" {
		return ""
	}
	if !strings.HasPrefix(t, "v") {
		t = "v" + t
	}
	t = semver.Canonical(t)
	if t == "" {
		return ""
	}
	if semver.Prerelease(t) != "" {
		return ""
	}
	return t
}

var latestVersionCache struct {
	mu        sync.Mutex
	version   string
	expiresAt time.Time
}

const listReleasesMaxPages = 10

// GetLatestVersion 获取最新版本（跳过 Helm chart 专用 tag：以 ech0- 开头）
func GetLatestVersion() (string, error) {
	now := time.Now().UTC()
	latestVersionCache.mu.Lock()
	if latestVersionCache.version != "" && now.Before(latestVersionCache.expiresAt) {
		v := latestVersionCache.version
		latestVersionCache.mu.Unlock()
		return v, nil
	}
	latestVersionCache.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client := github.NewClient(nil)
	owner, repo := "lin-snow", "Ech0"

	opts := &github.ListOptions{PerPage: 30, Page: 1}
	var best string
pageLoop:
	for page := 0; page < listReleasesMaxPages; page++ {
		releases, resp, err := client.Repositories.ListReleases(ctx, owner, repo, opts)
		if err != nil {
			return "", fmt.Errorf("list releases failed: %w", err)
		}
		for _, rel := range releases {
			if rel == nil {
				continue
			}
			if rel.GetDraft() {
				continue
			}
			tag := strings.TrimSpace(rel.GetTagName())
			if tag == "" {
				continue
			}
			if isHelmChartArtifactStyleTag(tag) {
				continue
			}
			if rel.GetPrerelease() {
				continue
			}
			canon := canonicalStableSemverFromReleaseTag(tag)
			if canon == "" {
				continue
			}
			best = canon
			break pageLoop
		}
		if resp == nil || resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	if best == "" {
		return "", fmt.Errorf("no stable application release found (ech0-* chart tags are ignored)")
	}

	// 保持与 versionPkg.Version 一致：返回不带 v 的 X.Y.Z
	result := strings.TrimPrefix(best, "v")

	latestVersionCache.mu.Lock()
	latestVersionCache.version = result
	latestVersionCache.expiresAt = time.Now().UTC().Add(30 * time.Minute)
	latestVersionCache.mu.Unlock()

	return result, nil
}
