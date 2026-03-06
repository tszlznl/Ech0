package local

import (
	"context"
	"io"
	"os"
	"path/filepath"

	stgx "github.com/lin-snow/ech0/pkg/storagex"
)

// LocalFS implements storagex.FS and storagex.URLResolver by
// mapping virtual paths onto a local directory tree.
//
//	root = "data/files"
//	virtual "/images/a.png" → physical "data/files/images/a.png"
type LocalFS struct {
	root      string
	urlPrefix string
}

type options struct {
	root      string
	urlPrefix string
}

func defaultOptions() options {
	return options{
		root:      "data/files",
		urlPrefix: "/files",
	}
}

// Option configures a LocalFS instance.
type Option func(*options)

// WithRoot sets the physical root directory for file storage.
func WithRoot(root string) Option {
	return func(o *options) { o.root = root }
}

// WithURLPrefix sets the URL prefix used by ResolveURL.
func WithURLPrefix(prefix string) Option {
	return func(o *options) { o.urlPrefix = prefix }
}

func NewLocalFS(opts ...Option) *LocalFS {
	o := defaultOptions()
	for _, fn := range opts {
		fn(&o)
	}
	return &LocalFS{
		root:      filepath.Clean(o.root),
		urlPrefix: o.urlPrefix,
	}
}

func (fs *LocalFS) physicalPath(virtualPath string) (string, error) {
	p, err := stgx.NormalizePath(virtualPath)
	if err != nil {
		return "", err
	}
	return filepath.Join(fs.root, stgx.TrimVirtualPath(p)), nil
}

func (fs *LocalFS) Open(_ context.Context, path string) (io.ReadCloser, error) {
	pp, err := fs.physicalPath(path)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(pp)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, stgx.ErrNotFound
		}
		return nil, err
	}
	return f, nil
}

func (fs *LocalFS) Write(_ context.Context, path string, r io.Reader, _ stgx.WriteOptions) error {
	pp, err := fs.physicalPath(path)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(pp), 0o750); err != nil {
		return err
	}
	f, err := os.Create(pp)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, r); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}

func (fs *LocalFS) Delete(_ context.Context, path string) error {
	pp, err := fs.physicalPath(path)
	if err != nil {
		return err
	}
	err = os.Remove(pp)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (fs *LocalFS) Stat(_ context.Context, path string) (*stgx.FileInfo, error) {
	pp, err := fs.physicalPath(path)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(pp)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, stgx.ErrNotFound
		}
		return nil, err
	}
	p, _ := stgx.NormalizePath(path)
	return &stgx.FileInfo{
		Path:    p,
		Size:    info.Size(),
		ModTime: info.ModTime(),
		IsDir:   info.IsDir(),
	}, nil
}

func (fs *LocalFS) List(_ context.Context, prefix string) ([]stgx.FileInfo, error) {
	p, err := stgx.NormalizePath(prefix)
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(fs.root, stgx.TrimVirtualPath(p))
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var result []stgx.FileInfo
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		result = append(result, stgx.FileInfo{
			Path:    stgx.JoinPath(p, e.Name()),
			Size:    info.Size(),
			ModTime: info.ModTime(),
			IsDir:   info.IsDir(),
		})
	}
	return result, nil
}

func (fs *LocalFS) Exists(_ context.Context, path string) (bool, error) {
	pp, err := fs.physicalPath(path)
	if err != nil {
		return false, nil
	}
	_, err = os.Stat(pp)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// ResolveURL implements storagex.URLResolver.
// Virtual "/images/a.png" → URL "/files/images/a.png".
func (fs *LocalFS) ResolveURL(_ context.Context, path string) (string, error) {
	p, err := stgx.NormalizePath(path)
	if err != nil {
		return "", err
	}
	return fs.urlPrefix + p, nil
}
