package storagex

import (
	"fmt"
	"path"
	"regexp"
	"strings"
)

var validSegment = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

// NormalizePath cleans and validates a virtual filesystem path.
// Result always starts with "/" and contains no "..", "//", or trailing "/".
func NormalizePath(p string) (string, error) {
	p = strings.TrimSpace(p)
	if p == "" {
		return "", ErrInvalidPath
	}
	// Reject traversal before path.Clean resolves it away
	if strings.Contains(p, "..") {
		return "", fmt.Errorf("%w: path traversal not allowed", ErrInvalidPath)
	}
	p = collapseSlashes(p)
	p = path.Clean(p)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	if p == "/" {
		return p, nil
	}
	segments := strings.Split(strings.TrimPrefix(p, "/"), "/")
	for _, seg := range segments {
		if seg == "" {
			continue
		}
		if !validSegment.MatchString(seg) {
			return "", fmt.Errorf("%w: invalid segment %q", ErrInvalidPath, seg)
		}
	}
	return p, nil
}

// ValidateSegment checks a single path segment for validity.
func ValidateSegment(seg string) error {
	seg = strings.TrimSpace(seg)
	if seg == "" || seg == "." || seg == ".." {
		return fmt.Errorf("%w: invalid segment %q", ErrInvalidPath, seg)
	}
	if !validSegment.MatchString(seg) {
		return fmt.Errorf("%w: invalid characters in segment %q", ErrInvalidPath, seg)
	}
	return nil
}

// JoinPath joins path segments with "/" and ensures a leading "/".
func JoinPath(segments ...string) string {
	joined := path.Join(segments...)
	if !strings.HasPrefix(joined, "/") {
		joined = "/" + joined
	}
	return joined
}

// TrimVirtualPath strips the leading "/" to produce a relative key
// suitable for object storage backends.
func TrimVirtualPath(p string) string {
	return strings.TrimPrefix(p, "/")
}

func collapseSlashes(input string) string {
	for strings.Contains(input, "//") {
		input = strings.ReplaceAll(input, "//", "/")
	}
	return input
}
