// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package util

import "testing"

func TestCanonicalStableSemverFromReleaseTag(t *testing.T) {
	tests := []struct {
		tag  string
		want string
	}{
		{"v4.5.0", "v4.5.0"},
		{"4.5.0", "v4.5.0"},
		{"ech0-4.5.0", ""},
		{"ECH0-4.5.0", ""},
		{"v4.5.0-rc.1", ""},
		{"not-a-version", ""},
		{"", ""},
	}
	for _, tt := range tests {
		got := canonicalStableSemverFromReleaseTag(tt.tag)
		if got != tt.want {
			t.Errorf("canonicalStableSemverFromReleaseTag(%q) = %q, want %q", tt.tag, got, tt.want)
		}
	}
}

func TestIsHelmChartArtifactStyleTag(t *testing.T) {
	tests := []struct {
		tag  string
		want bool
	}{
		{"ech0-4.5.0", true},
		{"ECH0-1.0.0", true},
		{"v4.5.0", false},
		{"ech0", false},
		{"", false},
	}
	for _, tt := range tests {
		got := isHelmChartArtifactStyleTag(tt.tag)
		if got != tt.want {
			t.Errorf("isHelmChartArtifactStyleTag(%q) = %v, want %v", tt.tag, got, tt.want)
		}
	}
}
