// Package s3x provides a minimal, SDK-agnostic interface for S3-compatible
// object storage and an AWS SDK v2 implementation.
package s3x

import (
	"context"
	"io"
	"time"
)

// Client defines the operations needed to interact with an S3-compatible
// object storage service. Method signatures use only standard library types
// so that callers never import the AWS SDK directly.
type Client interface {
	PutObject(ctx context.Context, bucket, key string, body io.Reader, contentType string) error
	GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error)
	DeleteObject(ctx context.Context, bucket, key string) error
	HeadObject(ctx context.Context, bucket, key string) (*ObjectInfo, error)
	ListObjects(ctx context.Context, bucket, prefix string) ([]ObjectEntry, error)
	PresignGetObject(ctx context.Context, bucket, key string, expiry time.Duration) (string, error)
	PresignPutObject(ctx context.Context, bucket, key string, expiry time.Duration) (string, error)
}

// ObjectInfo holds metadata returned by HeadObject.
type ObjectInfo struct {
	Key          string
	Size         int64
	ContentType  string
	LastModified time.Time
}

// ObjectEntry is one item returned by ListObjects.
type ObjectEntry struct {
	Key  string
	Size int64
}

// Config holds connection parameters for an S3-compatible service.
type Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Region    string
	UseSSL    bool
}
