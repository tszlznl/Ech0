package storage

import "context"

type StoragePort interface {
	Save(ctx context.Context, req SaveRequest) (FileObject, error)
	Delete(ctx context.Context, req DeleteRequest) error
	PresignUpload(ctx context.Context, req PresignRequest) (PresignResponse, error)
	ResolveURL(ctx context.Context, objectKey string) (string, error)
	Exists(ctx context.Context, path string) (bool, error)
}

type MetadataExtractorPort interface {
	Extract(ctx context.Context, category Category, contentType string, reader ReadSeekCloser) (FileMetadata, error)
}

