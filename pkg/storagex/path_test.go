package storagex

import (
	"testing"
)

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"simple", "/images/a.png", "/images/a.png", false},
		{"no leading slash", "images/a.png", "/images/a.png", false},
		{"double slash", "/images//a.png", "/images/a.png", false},
		{"trailing slash", "/images/", "/images", false},
		{"traversal", "/images/../etc/passwd", "", true},
		{"empty", "", "", true},
		{"special chars", "/images/a b.png", "", true},
		{"root", "/", "/", false},
		{"nested", "/a/b/c.txt", "/a/b/c.txt", false},
		{"dot in name", "/images/photo.2024.jpg", "/images/photo.2024.jpg", false},
		{"hyphen and underscore", "/dir-name/file_name.txt", "/dir-name/file_name.txt", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizePath(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NormalizePath(%q) error=%v, wantErr=%v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Fatalf("NormalizePath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidateSegment(t *testing.T) {
	tests := []struct {
		seg     string
		wantErr bool
	}{
		{"images", false},
		{"photo.png", false},
		{"my-file_01", false},
		{"..", true},
		{".", true},
		{"", true},
		{"a b", true},
		{"a/b", true},
	}
	for _, tt := range tests {
		t.Run(tt.seg, func(t *testing.T) {
			err := ValidateSegment(tt.seg)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateSegment(%q) error=%v, wantErr=%v", tt.seg, err, tt.wantErr)
			}
		})
	}
}

func TestJoinPath(t *testing.T) {
	got := JoinPath("images", "photo.png")
	if got != "/images/photo.png" {
		t.Fatalf("JoinPath = %q, want /images/photo.png", got)
	}
}

func TestTrimVirtualPath(t *testing.T) {
	got := TrimVirtualPath("/images/photo.png")
	if got != "images/photo.png" {
		t.Fatalf("TrimVirtualPath = %q, want images/photo.png", got)
	}
}
