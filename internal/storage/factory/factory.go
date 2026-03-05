package factory

import (
	"fmt"
	"strings"

	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	"github.com/lin-snow/ech0/internal/storage"
	localStorage "github.com/lin-snow/ech0/internal/storage/local"
	s3Storage "github.com/lin-snow/ech0/internal/storage/s3"
	"github.com/spf13/afero"
)

type Mode string

const (
	ModeLocal Mode = "local"
	ModeS3    Mode = "s3"
)

type BuildInput struct {
	Mode      Mode
	FS        afero.Fs
	S3Setting settingModel.S3Setting
}

func Build(input BuildInput) (storage.StoragePort, error) {
	switch Mode(strings.ToLower(strings.TrimSpace(string(input.Mode)))) {
	case ModeS3:
		if !input.S3Setting.Enable {
			return nil, fmt.Errorf("s3 mode selected but s3 is disabled")
		}
		return s3Storage.NewAdapter(input.S3Setting)
	case ModeLocal, "":
		return localStorage.NewAdapter(input.FS), nil
	default:
		return nil, fmt.Errorf("unsupported storage mode: %s", input.Mode)
	}
}

