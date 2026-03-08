package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	virefs "github.com/lin-snow/VireFS"
	"github.com/lin-snow/ech0/internal/config"
)

type StorageSelector struct {
	localFS        virefs.FS
	objectFS       virefs.FS
	localResolve   URLResolver
	objectResolve  URLResolver
	objectEnabled  bool
	objectProvider string
	objectBucket   string
}

func NewStorageSelector(cfg config.StorageConfig) *StorageSelector {
	schema := NewFileSchema()

	localFS := buildLocalFS(cfg, schema)
	localResolve := buildLocalURLResolver(schema)

	objectFS, objectResolve, objectEnabled := buildOptionalObjectFSAndResolver(cfg, schema)

	return &StorageSelector{
		localFS:        localFS,
		objectFS:       objectFS,
		localResolve:   localResolve,
		objectResolve:  objectResolve,
		objectEnabled:  objectEnabled,
		objectProvider: strings.ToLower(strings.TrimSpace(cfg.Provider)),
		objectBucket:   strings.TrimSpace(cfg.BucketName),
	}
}

func (r *StorageSelector) ObjectEnabled() bool {
	return r != nil && r.objectEnabled && r.objectFS != nil
}

func (r *StorageSelector) ObjectRoute() (provider string, bucket string) {
	if r == nil || !r.ObjectEnabled() {
		return "", ""
	}
	return r.objectProvider, r.objectBucket
}

func (r *StorageSelector) Put(
	ctx context.Context,
	storageType StorageType,
	key string,
	reader io.Reader,
	opts ...virefs.PutOption,
) error {
	fs, err := r.getFS(storageType)
	if err != nil {
		return err
	}
	return fs.Put(ctx, key, reader, opts...)
}

func (r *StorageSelector) Get(ctx context.Context, storageType StorageType, key string) (io.ReadCloser, error) {
	fs, err := r.getFS(storageType)
	if err != nil {
		return nil, err
	}
	return fs.Get(ctx, key)
}

func (r *StorageSelector) Delete(ctx context.Context, storageType StorageType, key string) error {
	fs, err := r.getFS(storageType)
	if err != nil {
		return err
	}
	return fs.Delete(ctx, key)
}

func (r *StorageSelector) ResolveURL(storageType StorageType, key string) string {
	if r == nil {
		return ""
	}
	switch NormalizeStorageType(string(storageType)) {
	case StorageTypeObject:
		if r.ObjectEnabled() && r.objectResolve != nil {
			return r.objectResolve(key)
		}
		return ""
	default:
		if r.localResolve != nil {
			return r.localResolve(key)
		}
		return ""
	}
}

func (r *StorageSelector) PresignPutURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	if !r.ObjectEnabled() {
		return "", errors.New("backend does not support presigned URLs")
	}
	p, ok := r.objectFS.(virefs.Presigner)
	if !ok {
		return "", errors.New("backend does not support presigned URLs")
	}
	req, err := p.PresignPut(ctx, key, expires)
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

func (r *StorageSelector) getFS(storageType StorageType) (virefs.FS, error) {
	if r == nil {
		return nil, errors.New("storage selector is not initialized")
	}
	switch NormalizeStorageType(string(storageType)) {
	case StorageTypeObject:
		if !r.ObjectEnabled() {
			return nil, errors.New("object storage is not enabled")
		}
		return r.objectFS, nil
	case StorageTypeExternal:
		return nil, errors.New("external storage does not support filesystem operations")
	default:
		return r.localFS, nil
	}
}

func buildOptionalObjectFSAndResolver(
	cfg config.StorageConfig,
	schema *virefs.Schema,
) (virefs.FS, URLResolver, bool) {
	mode := strings.ToLower(strings.TrimSpace(cfg.Mode))
	if !cfg.ObjectEnabled && NormalizeStorageMode(mode) != StorageModeObject {
		return nil, nil, false
	}

	provider := mapProvider(cfg.Provider)
	region := resolveObjectRegion(cfg.Provider, cfg.Region)
	var opts []virefs.ObjectOption
	if cfg.PathPrefix != "" {
		opts = append(opts, virefs.WithPrefix(strings.Trim(cfg.PathPrefix, "/")+"/"))
	}
	opts = append(opts, virefs.WithObjectKeyFunc(schema.Resolve))

	endpoint := normalizeEndpoint(cfg.Endpoint, cfg.UseSSL)
	fs, err := virefs.NewObjectFSFromConfig(context.Background(), &virefs.S3Config{
		Provider:  provider,
		Endpoint:  endpoint,
		Region:    region,
		Bucket:    cfg.BucketName,
		AccessKey: cfg.AccessKey,
		SecretKey: cfg.SecretKey,
	}, opts...)
	if err != nil {
		return nil, nil, false
	}

	return fs, buildS3URLResolver(cfg, schema), true
}

func (r *StorageSelector) CapabilityText() string {
	if r.ObjectEnabled() {
		return "local+object"
	}
	return "local"
}

func (r *StorageSelector) String() string {
	return fmt.Sprintf("StorageSelector(%s)", r.CapabilityText())
}
