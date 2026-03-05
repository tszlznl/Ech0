package local

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/lin-snow/ech0/internal/config"
	"github.com/lin-snow/ech0/internal/storage"
	fileUtil "github.com/lin-snow/ech0/internal/util/file"
	storageUtil "github.com/lin-snow/ech0/internal/util/storage"
	"github.com/spf13/afero"
)

type Adapter struct {
	fs afero.Fs
}

func NewAdapter(fs afero.Fs) *Adapter {
	return &Adapter{fs: fs}
}

func (a *Adapter) Save(_ context.Context, req storage.SaveRequest) (storage.FileObject, error) {
	dir, prefix := targetDirAndURLPrefix(req.Category)
	if err := a.fs.MkdirAll(dir, 0o750); err != nil {
		return storage.FileObject{}, err
	}

	ext := filepath.Ext(req.FileName)
	fileName, err := storageUtil.GenerateRandomFilename(req.UserID, ext)
	if err != nil {
		return storage.FileObject{}, err
	}
	if req.Category == storage.CategoryAudio {
		fileName = "music" + ext
	}
	savePath := filepath.Join(dir, fileName)
	out, err := a.fs.Create(savePath)
	if err != nil {
		return storage.FileObject{}, err
	}
	if _, err := io.Copy(out, req.Reader); err != nil {
		_ = out.Close()
		return storage.FileObject{}, err
	}
	if err := out.Close(); err != nil {
		return storage.FileObject{}, err
	}

	url := fmt.Sprintf("%s%s", prefix, fileName)
	return storage.FileObject{
		URL:         url,
		Source:      storage.SourceLocal,
		ContentType: req.ContentType,
		Category:    req.Category,
	}, nil
}

func (a *Adapter) Delete(_ context.Context, req storage.DeleteRequest) error {
	if req.URL == "" {
		return nil
	}
	baseDir, prefix := resolveBaseDirAndPrefix(req)
	localPath, err := fileUtil.ValidateAndSanitizePath(baseDir, req.URL, prefix)
	if err != nil {
		return err
	}
	return storageUtil.DeleteFileFromLocal(a.fs, localPath)
}

func (a *Adapter) PresignUpload(_ context.Context, _ storage.PresignRequest) (storage.PresignResponse, error) {
	return storage.PresignResponse{}, fmt.Errorf("local storage does not support presign upload")
}

func (a *Adapter) ResolveURL(_ context.Context, objectKey string) (string, error) {
	if objectKey == "" {
		return "", fmt.Errorf("object key is empty")
	}
	return objectKey, nil
}

func (a *Adapter) Exists(_ context.Context, path string) (bool, error) {
	_, err := a.fs.Stat(path)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func resolveBaseDirAndPrefix(req storage.DeleteRequest) (string, string) {
	if req.Category == storage.CategoryAudio || strings.HasPrefix(req.URL, "/audios/") {
		return filepath.Clean(config.Config().Upload.AudioPath), "/audios/"
	}
	return filepath.Clean(config.Config().Upload.ImagePath), "/images/"
}

func targetDirAndURLPrefix(category storage.Category) (string, string) {
	if category == storage.CategoryAudio {
		return filepath.Clean(config.Config().Upload.AudioPath), "/audios/"
	}
	return filepath.Clean(config.Config().Upload.ImagePath), "/images/"
}

