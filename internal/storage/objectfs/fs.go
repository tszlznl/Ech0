package objectfs

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/lin-snow/ech0/pkg/s3x"
	stgx "github.com/lin-snow/ech0/pkg/storagex"
)

// ObjectFS implements storagex.FS, storagex.URLResolver and storagex.Signer
// by mapping virtual paths onto S3-compatible object storage keys.
//
//	pathPrefix = "uploads"
//	virtual "/images/a.png" → key "uploads/images/a.png"
type ObjectFS struct {
	client     s3x.Client
	cfg        stgx.ObjectStorageConfig
	bucket     string
	pathPrefix string
}

type options struct {
	pathPrefix string
}

// Option configures an ObjectFS instance.
type Option func(*options)

// WithPathPrefix overrides the object key prefix inside the bucket.
func WithPathPrefix(prefix string) Option {
	return func(o *options) { o.pathPrefix = prefix }
}

// New creates an ObjectFS backed by the given s3x.Client.
// The client is created externally (e.g. via s3x.New) and injected here,
// keeping ObjectFS free of SDK construction logic.
func New(client s3x.Client, cfg stgx.ObjectStorageConfig, opts ...Option) *ObjectFS {
	o := options{pathPrefix: strings.Trim(cfg.PathPrefix, "/")}
	for _, fn := range opts {
		fn(&o)
	}
	return &ObjectFS{
		client:     client,
		cfg:        cfg,
		bucket:     cfg.BucketName,
		pathPrefix: o.pathPrefix,
	}
}

func (fs *ObjectFS) objectKey(virtualPath string) (string, error) {
	p, err := stgx.NormalizePath(virtualPath)
	if err != nil {
		return "", err
	}
	relative := stgx.TrimVirtualPath(p)
	if fs.pathPrefix != "" {
		return fs.pathPrefix + "/" + relative, nil
	}
	return relative, nil
}

func (fs *ObjectFS) Open(ctx context.Context, path string) (io.ReadCloser, error) {
	key, err := fs.objectKey(path)
	if err != nil {
		return nil, err
	}
	return fs.client.GetObject(ctx, fs.bucket, key)
}

func (fs *ObjectFS) Write(ctx context.Context, path string, r io.Reader, opts stgx.WriteOptions) error {
	key, err := fs.objectKey(path)
	if err != nil {
		return err
	}
	return fs.client.PutObject(ctx, fs.bucket, key, r, opts.ContentType)
}

func (fs *ObjectFS) Delete(ctx context.Context, path string) error {
	key, err := fs.objectKey(path)
	if err != nil {
		return err
	}
	return fs.client.DeleteObject(ctx, fs.bucket, key)
}

func (fs *ObjectFS) Stat(ctx context.Context, path string) (*stgx.FileInfo, error) {
	key, err := fs.objectKey(path)
	if err != nil {
		return nil, err
	}
	info, err := fs.client.HeadObject(ctx, fs.bucket, key)
	if err != nil {
		return nil, stgx.ErrNotFound
	}
	p, _ := stgx.NormalizePath(path)
	return &stgx.FileInfo{
		Path:        p,
		Size:        info.Size,
		ContentType: info.ContentType,
		ModTime:     info.LastModified,
	}, nil
}

func (fs *ObjectFS) List(ctx context.Context, prefix string) ([]stgx.FileInfo, error) {
	key, err := fs.objectKey(prefix)
	if err != nil {
		return nil, err
	}
	if !strings.HasSuffix(key, "/") {
		key += "/"
	}
	entries, err := fs.client.ListObjects(ctx, fs.bucket, key)
	if err != nil {
		return nil, err
	}
	var result []stgx.FileInfo
	for _, e := range entries {
		result = append(result, stgx.FileInfo{
			Path: fs.keyToVirtualPath(e.Key),
			Size: e.Size,
		})
	}
	return result, nil
}

func (fs *ObjectFS) Exists(ctx context.Context, path string) (bool, error) {
	key, err := fs.objectKey(path)
	if err != nil {
		return false, nil
	}
	_, err = fs.client.HeadObject(ctx, fs.bucket, key)
	return err == nil, nil
}

// ResolveURL implements storagex.URLResolver.
func (fs *ObjectFS) ResolveURL(_ context.Context, path string) (string, error) {
	key, err := fs.objectKey(path)
	if err != nil {
		return "", err
	}
	protocol := "http"
	if fs.cfg.UseSSL {
		protocol = "https"
	}
	baseURL := fmt.Sprintf("%s://%s/%s", protocol, fs.cfg.Endpoint, fs.cfg.BucketName)
	if cdn := strings.TrimSpace(fs.cfg.CDNURL); cdn != "" {
		cdnURL := strings.TrimRight(cdn, "/")
		if !strings.HasPrefix(strings.ToLower(cdnURL), "http://") && !strings.HasPrefix(strings.ToLower(cdnURL), "https://") {
			cdnURL = fmt.Sprintf("%s://%s", protocol, cdnURL)
		}
		baseURL = cdnURL
	}
	return fmt.Sprintf("%s/%s", baseURL, strings.TrimLeft(key, "/")), nil
}

// Sign implements storagex.Signer.
func (fs *ObjectFS) Sign(ctx context.Context, path string, method string, expiry time.Duration) (string, error) {
	key, err := fs.objectKey(path)
	if err != nil {
		return "", err
	}
	switch strings.ToUpper(method) {
	case "GET":
		return fs.client.PresignGetObject(ctx, fs.bucket, key, expiry)
	case "PUT":
		return fs.client.PresignPutObject(ctx, fs.bucket, key, expiry)
	default:
		return "", fmt.Errorf("unsupported presign method: %s", method)
	}
}

func (fs *ObjectFS) keyToVirtualPath(key string) string {
	key = strings.TrimLeft(key, "/")
	if fs.pathPrefix != "" {
		key = strings.TrimPrefix(key, fs.pathPrefix+"/")
	}
	return "/" + key
}
