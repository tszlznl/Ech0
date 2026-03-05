package s3

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	"github.com/lin-snow/ech0/internal/storage"
	storageUtil "github.com/lin-snow/ech0/internal/util/storage"
)

type Adapter struct {
	client  storageUtil.ObjectStorage
	setting settingModel.S3Setting
}

func NewAdapter(setting settingModel.S3Setting) (*Adapter, error) {
	client, err := storageUtil.NewMinioStorage(
		setting.Endpoint,
		setting.AccessKey,
		setting.SecretKey,
		setting.BucketName,
		setting.Region,
		setting.Provider,
		setting.UseSSL,
	)
	if err != nil {
		return nil, err
	}
	return &Adapter{client: client, setting: setting}, nil
}

func (a *Adapter) Save(ctx context.Context, req storage.SaveRequest) (storage.FileObject, error) {
	objectKey, err := buildObjectKey(req.UserID, req.FileName, a.setting.PathPrefix)
	if err != nil {
		return storage.FileObject{}, err
	}
	if _, err := req.Reader.Seek(0, io.SeekStart); err != nil {
		return storage.FileObject{}, err
	}
	if err := a.client.Upload(ctx, objectKey, req.Reader, req.ContentType); err != nil {
		return storage.FileObject{}, err
	}
	url, err := a.ResolveURL(ctx, objectKey)
	if err != nil {
		return storage.FileObject{}, err
	}
	return storage.FileObject{
		URL:         url,
		Source:      storage.SourceS3,
		ObjectKey:   objectKey,
		ContentType: req.ContentType,
		Category:    req.Category,
	}, nil
}

func (a *Adapter) Delete(ctx context.Context, req storage.DeleteRequest) error {
	if req.ObjectKey == "" {
		return nil
	}
	return a.client.DeleteObject(ctx, req.ObjectKey)
}

func (a *Adapter) PresignUpload(ctx context.Context, req storage.PresignRequest) (storage.PresignResponse, error) {
	exp := req.Expiry
	if exp <= 0 {
		exp = 24 * time.Hour
	}
	objectKey, err := buildObjectKey(req.UserID, req.FileName, a.setting.PathPrefix)
	if err != nil {
		return storage.PresignResponse{}, err
	}
	presignURL, err := a.client.PresignURL(ctx, objectKey, exp, req.Method)
	if err != nil {
		return storage.PresignResponse{}, err
	}
	fileURL, err := a.ResolveURL(ctx, objectKey)
	if err != nil {
		return storage.PresignResponse{}, err
	}
	return storage.PresignResponse{
		FileName:    req.FileName,
		ContentType: req.ContentType,
		ObjectKey:   objectKey,
		PresignURL:  presignURL,
		FileURL:     fileURL,
	}, nil
}

func (a *Adapter) ResolveURL(_ context.Context, objectKey string) (string, error) {
	if objectKey == "" {
		return "", fmt.Errorf("object key is empty")
	}
	protocol := "http"
	if a.setting.UseSSL {
		protocol = "https"
	}
	baseURL := fmt.Sprintf("%s://%s/%s", protocol, a.setting.Endpoint, a.setting.BucketName)
	if trimmedCDN := strings.TrimSpace(a.setting.CDNURL); trimmedCDN != "" {
		cdnURL := strings.TrimRight(trimmedCDN, "/")
		if !strings.HasPrefix(strings.ToLower(cdnURL), "http://") && !strings.HasPrefix(strings.ToLower(cdnURL), "https://") {
			cdnURL = fmt.Sprintf("%s://%s", protocol, cdnURL)
		}
		baseURL = cdnURL
	}
	return fmt.Sprintf("%s/%s", baseURL, strings.TrimLeft(objectKey, "/")), nil
}

func (a *Adapter) Exists(context.Context, string) (bool, error) {
	return true, nil
}

func buildObjectKey(userID uint, fileName, prefix string) (string, error) {
	prefix = strings.Trim(prefix, "/")
	ext := filepath.Ext(fileName)
	if ext == "" {
		ext = ".bin"
	}
	key, err := storageUtil.GenerateRandomFilename(userID, ext)
	if err != nil {
		return "", err
	}
	if prefix == "" {
		return key, nil
	}
	return strings.TrimPrefix(prefix+"/"+key, "/"), nil
}

