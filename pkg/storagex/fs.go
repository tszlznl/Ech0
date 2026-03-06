package storagex

import (
	"context"
	"io"
	"time"
)

// FS is the unified virtual filesystem interface.
// Business code operates on virtual paths (e.g. /images/a.png)
// without knowing whether the backend is local disk or S3.
type FS interface {
	Open(ctx context.Context, path string) (io.ReadCloser, error)
	Write(ctx context.Context, path string, r io.Reader, opts WriteOptions) error
	Delete(ctx context.Context, path string) error
	Stat(ctx context.Context, path string) (*FileInfo, error)
	List(ctx context.Context, prefix string) ([]FileInfo, error)
	Exists(ctx context.Context, path string) (bool, error)
}

// Signer is an optional capability for backends that support presigned URLs.
type Signer interface {
	Sign(ctx context.Context, path string, method string, expiry time.Duration) (string, error)
}

// URLResolver resolves a virtual path to a publicly accessible URL.
type URLResolver interface {
	ResolveURL(ctx context.Context, path string) (string, error)
}

// FileInfo describes a file or directory in the virtual filesystem.
type FileInfo struct {
	Path        string    `json:"path"`
	Size        int64     `json:"size"`
	ContentType string    `json:"content_type,omitempty"`
	ModTime     time.Time `json:"mod_time,omitempty"`
	IsDir       bool      `json:"is_dir,omitempty"`
}

// WriteOptions controls how a file is written.
type WriteOptions struct {
	ContentType string
}
