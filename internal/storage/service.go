package storage

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	stgx "github.com/lin-snow/ech0/pkg/storagex"
)

// StorageService provides business-level file operations on top of the
// unified VFS. It replaces the old StoragePort interface and PortBridge
// adapter, giving callers direct access to virtual-path-based operations.
type StorageService struct {
	fs     stgx.FS
	keyGen stgx.KeyGenerator
	source string // "local" or "s3" — DB compat label
}

type StorageServiceConfig struct {
	FS     stgx.FS
	KeyGen stgx.KeyGenerator
	Source string
}

func NewStorageService(cfg StorageServiceConfig) *StorageService {
	keyGen := cfg.KeyGen
	if keyGen == nil {
		keyGen = stgx.NewRandomKeyGenerator()
	}
	return &StorageService{
		fs:     cfg.FS,
		keyGen: keyGen,
		source: cfg.Source,
	}
}

// UploadResult is the outcome of a successful Upload call.
type UploadResult struct {
	VirtualPath string
	URL         string
	ObjectKey   string
	ContentType string
}

// PresignResult is the outcome of a successful Presign call.
type PresignResult struct {
	VirtualPath string
	ObjectKey   string
	PresignURL  string
	FileURL     string
}

// Upload generates a unique virtual path, writes the file, and resolves
// the public URL. Audio category files use a static "music.ext" path.
func (s *StorageService) Upload(
	ctx context.Context,
	category stgx.Category,
	userID uint,
	fileName string,
	contentType string,
	r io.Reader,
) (UploadResult, error) {
	gen := s.keyGenForCategory(category, fileName)
	vpath, err := gen.GenerateKey(category, userID, fileName)
	if err != nil {
		return UploadResult{}, err
	}

	if seeker, ok := r.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return UploadResult{}, err
		}
	}

	if err := s.fs.Write(ctx, vpath, r, stgx.WriteOptions{ContentType: contentType}); err != nil {
		return UploadResult{}, err
	}

	url := s.resolveURLOrFallback(ctx, vpath)

	return UploadResult{
		VirtualPath: vpath,
		URL:         url,
		ObjectKey:   stgx.TrimVirtualPath(vpath),
		ContentType: contentType,
	}, nil
}

// Delete removes the file at the given virtual path.
func (s *StorageService) Delete(ctx context.Context, path string) error {
	return s.fs.Delete(ctx, path)
}

// Presign generates a unique virtual path and returns a presigned URL
// for client-side uploads. Only works when the backend implements Signer.
func (s *StorageService) Presign(
	ctx context.Context,
	category stgx.Category,
	userID uint,
	fileName string,
	contentType string,
	method string,
	expiry time.Duration,
) (PresignResult, error) {
	signer, ok := s.fs.(stgx.Signer)
	if !ok {
		return PresignResult{}, fmt.Errorf("backend does not support presigned URLs")
	}

	gen := s.keyGenForCategory(category, fileName)
	vpath, err := gen.GenerateKey(category, userID, fileName)
	if err != nil {
		return PresignResult{}, err
	}

	if expiry <= 0 {
		expiry = 24 * time.Hour
	}

	presignURL, err := signer.Sign(ctx, vpath, method, expiry)
	if err != nil {
		return PresignResult{}, err
	}

	fileURL := s.resolveURLOrFallback(ctx, vpath)

	return PresignResult{
		VirtualPath: vpath,
		ObjectKey:   stgx.TrimVirtualPath(vpath),
		PresignURL:  presignURL,
		FileURL:     fileURL,
	}, nil
}

// ResolveURL maps a virtual path to a publicly accessible URL.
func (s *StorageService) ResolveURL(ctx context.Context, virtualPath string) (string, error) {
	if virtualPath == "" {
		return "", fmt.Errorf("virtual path is empty")
	}
	resolver, ok := s.fs.(stgx.URLResolver)
	if !ok {
		return virtualPath, nil
	}
	return resolver.ResolveURL(ctx, virtualPath)
}

// Open returns a reader for the file at the given virtual path.
func (s *StorageService) Open(ctx context.Context, path string) (io.ReadCloser, error) {
	return s.fs.Open(ctx, path)
}

// Stat returns metadata for the file at the given virtual path.
func (s *StorageService) Stat(ctx context.Context, path string) (*stgx.FileInfo, error) {
	return s.fs.Stat(ctx, path)
}

// Exists checks whether a file exists at the given virtual path.
func (s *StorageService) Exists(ctx context.Context, path string) (bool, error) {
	return s.fs.Exists(ctx, path)
}

// Source returns the backend label ("local" or "s3") for DB compatibility.
func (s *StorageService) Source() string {
	return s.source
}

// VFS exposes the underlying storagex.FS for advanced operations
// like migration or direct streaming.
func (s *StorageService) VFS() stgx.FS {
	return s.fs
}

func (s *StorageService) keyGenForCategory(category stgx.Category, fileName string) stgx.KeyGenerator {
	if category == stgx.CategoryAudio {
		ext := strings.ToLower(filepath.Ext(strings.TrimSpace(fileName)))
		if ext == "" {
			ext = ".bin"
		}
		return &stgx.StaticKeyGenerator{
			Category: stgx.CategoryAudio,
			Name:     "music" + ext,
		}
	}
	return s.keyGen
}

func (s *StorageService) resolveURLOrFallback(ctx context.Context, vpath string) string {
	if resolver, ok := s.fs.(stgx.URLResolver); ok {
		if resolved, err := resolver.ResolveURL(ctx, vpath); err == nil {
			return resolved
		}
	}
	return vpath
}
