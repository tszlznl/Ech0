package factory

import (
	"fmt"
	"strings"

	"github.com/lin-snow/ech0/internal/storage"
	localStorage "github.com/lin-snow/ech0/internal/storage/local"
	objectStorage "github.com/lin-snow/ech0/internal/storage/objectfs"
	stgx "github.com/lin-snow/ech0/pkg/storagex"
)

type Mode string

const (
	ModeLocal Mode = "local"
	ModeS3    Mode = "s3"
)

type BuildInput struct {
	Mode         Mode
	DataRoot     string                    // local filesystem root, e.g. "data/files"
	ObjectConfig *stgx.ObjectStorageConfig // object storage config (S3/R2/MinIO/...)
}

// Build creates a StorageService backed by the unified VFS.
func Build(input BuildInput) (*storage.StorageService, error) {
	switch Mode(strings.ToLower(strings.TrimSpace(string(input.Mode)))) {
	case ModeS3:
		if input.ObjectConfig == nil {
			return nil, fmt.Errorf("s3 mode selected but no object storage config provided")
		}
		return buildObject(*input.ObjectConfig)
	case ModeLocal, "":
		return buildLocal(input.DataRoot)
	default:
		return nil, fmt.Errorf("unsupported storage mode: %s", input.Mode)
	}
}

// BuildFS creates a raw storagex.FS without the StorageService wrapper.
func BuildFS(input BuildInput) (stgx.FS, error) {
	switch Mode(strings.ToLower(strings.TrimSpace(string(input.Mode)))) {
	case ModeS3:
		if input.ObjectConfig == nil {
			return nil, fmt.Errorf("s3 mode selected but no object storage config provided")
		}
		return objectStorage.New(*input.ObjectConfig)
	case ModeLocal, "":
		var opts []localStorage.Option
		if input.DataRoot != "" {
			opts = append(opts, localStorage.WithRoot(input.DataRoot))
		}
		return localStorage.NewLocalFS(opts...), nil
	default:
		return nil, fmt.Errorf("unsupported storage mode: %s", input.Mode)
	}
}

func buildLocal(dataRoot string) (*storage.StorageService, error) {
	var opts []localStorage.Option
	if dataRoot != "" {
		opts = append(opts, localStorage.WithRoot(dataRoot))
	}
	fs := localStorage.NewLocalFS(opts...)
	return storage.NewStorageService(storage.StorageServiceConfig{
		FS:     fs,
		Source: string(storage.SourceLocal),
	}), nil
}

func buildObject(cfg stgx.ObjectStorageConfig) (*storage.StorageService, error) {
	fs, err := objectStorage.New(cfg)
	if err != nil {
		return nil, err
	}
	return storage.NewStorageService(storage.StorageServiceConfig{
		FS:     fs,
		Source: string(storage.SourceS3),
	}), nil
}
