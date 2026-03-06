package objectfs

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	storageUtil "github.com/lin-snow/ech0/internal/util/storage"
	stgx "github.com/lin-snow/ech0/pkg/storagex"
)

// ObjectFS implements storagex.FS, storagex.URLResolver and storagex.Signer
// by mapping virtual paths onto S3-compatible object storage keys.
//
//	pathPrefix = "uploads"
//	virtual "/images/a.png" → key "uploads/images/a.png"
type ObjectFS struct {
	client     storageUtil.ObjectStorage
	cfg        stgx.ObjectStorageConfig
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

func New(cfg stgx.ObjectStorageConfig, opts ...Option) (*ObjectFS, error) {
	o := options{pathPrefix: strings.Trim(cfg.PathPrefix, "/")}
	for _, fn := range opts {
		fn(&o)
	}

	client, err := storageUtil.NewMinioStorage(
		cfg.Endpoint,
		cfg.AccessKey,
		cfg.SecretKey,
		cfg.BucketName,
		cfg.Region,
		cfg.Provider,
		cfg.UseSSL,
	)
	if err != nil {
		return nil, err
	}
	return &ObjectFS{
		client:     client,
		cfg:        cfg,
		pathPrefix: o.pathPrefix,
	}, nil
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
	return fs.client.Download(ctx, key)
}

func (fs *ObjectFS) Write(ctx context.Context, path string, r io.Reader, opts stgx.WriteOptions) error {
	key, err := fs.objectKey(path)
	if err != nil {
		return err
	}
	return fs.client.Upload(ctx, key, r, opts.ContentType)
}

func (fs *ObjectFS) Delete(ctx context.Context, path string) error {
	key, err := fs.objectKey(path)
	if err != nil {
		return err
	}
	return fs.client.DeleteObject(ctx, key)
}

func (fs *ObjectFS) Stat(ctx context.Context, path string) (*stgx.FileInfo, error) {
	key, err := fs.objectKey(path)
	if err != nil {
		return nil, err
	}
	objects, err := fs.client.ListObjects(ctx, key)
	if err != nil {
		return nil, err
	}
	for _, k := range objects {
		if strings.TrimLeft(k, "/") == key {
			p, _ := stgx.NormalizePath(path)
			return &stgx.FileInfo{Path: p}, nil
		}
	}
	return nil, stgx.ErrNotFound
}

func (fs *ObjectFS) List(ctx context.Context, prefix string) ([]stgx.FileInfo, error) {
	key, err := fs.objectKey(prefix)
	if err != nil {
		return nil, err
	}
	objects, err := fs.client.ListObjects(ctx, key)
	if err != nil {
		return nil, err
	}
	var result []stgx.FileInfo
	for _, k := range objects {
		result = append(result, stgx.FileInfo{Path: fs.keyToVirtualPath(k)})
	}
	return result, nil
}

func (fs *ObjectFS) Exists(ctx context.Context, path string) (bool, error) {
	key, err := fs.objectKey(path)
	if err != nil {
		return false, nil
	}
	objects, err := fs.client.ListObjects(ctx, key)
	if err != nil {
		return false, err
	}
	for _, k := range objects {
		if strings.TrimLeft(k, "/") == key {
			return true, nil
		}
	}
	return false, nil
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
	return fs.client.PresignURL(ctx, key, expiry, method)
}

func (fs *ObjectFS) keyToVirtualPath(key string) string {
	key = strings.TrimLeft(key, "/")
	if fs.pathPrefix != "" {
		key = strings.TrimPrefix(key, fs.pathPrefix+"/")
	}
	return "/" + key
}
