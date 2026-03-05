package util

import (
	"context"
	"errors"
	"io"
	"mime/multipart"
	"time"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	"github.com/spf13/afero"
)

// UploadFile 根据文件类型和存储类型上传文件
func UploadFile(
	fs afero.Fs,
	file *multipart.FileHeader,
	fileType commonModel.UploadFileType,
	storageType commonModel.FileStorageType,
	userID uint,
) (string, error) {
	if file == nil {
		return "", errors.New(commonModel.NO_FILE_UPLOAD_ERROR)
	}

	switch storageType {
	case commonModel.LOCAL_FILE:
		return UploadFileToLocal(fs, file, fileType, userID)
	case commonModel.S3_FILE:
		// TODO: Implement S3 file upload
	default:
		return "", errors.New(commonModel.NO_FILE_STORAGE_ERROR)
	}

	return "", errors.New(commonModel.NO_FILE_STORAGE_ERROR)
}

// IsAllowedType 检查Content-Type是否在允许的类型列表中
func IsAllowedType(contentType string, allowedTypes []string) bool {
	for _, allowed := range allowedTypes {
		if contentType == allowed {
			return true
		}
	}
	return false
}

// createDirIfNotExist 创建目录如果不存在
func createDirIfNotExist(fs afero.Fs, imagePath string) error {
	if _, err := fs.Stat(imagePath); err != nil {
		if err := fs.MkdirAll(imagePath, 0o755); err != nil {
			return err
		}
	}
	return nil
}

// FileExists 文件是否存在
func FileExists(fs afero.Fs, filePath string) bool {
	_, err := fs.Stat(filePath)
	return err == nil
}

// ObjectStorage 对象存储接口
type ObjectStorage interface {
	// Upload 上传文件到对象存储
	Upload(ctx context.Context, objectName string, r io.Reader, contentType string) error

	// Download 下载对象存储中的文件
	Download(ctx context.Context, objectName string) (io.ReadCloser, error)

	// ListObjects 列出对象存储中的文件
	ListObjects(ctx context.Context, prefix string) ([]string, error)

	// ListObjectStream 列出对象存储中的文件流
	ListObjectStream(ctx context.Context, prefix string) (<-chan string, error)

	// DeleteObject 删除对象存储中的文件
	DeleteObject(ctx context.Context, objectName string) error

	// PresignURL 生成对象存储中文件的临时访问链接
	PresignURL(
		ctx context.Context,
		objectName string,
		expiry time.Duration,
		method string,
	) (string, error)
}

// FileStoragePort 提供统一的文件存储能力（local/s3 对齐）
type FileStoragePort interface {
	UploadLocal(file *multipart.FileHeader, fileType commonModel.UploadFileType, userID uint) (string, error)
	DeleteLocal(filePath string) error
	PresignURL(ctx context.Context, userID uint, dto *commonModel.GetPresignURLDto, method string, setting settingModel.S3Setting) (commonModel.PresignDto, error)
	DeleteObject(ctx context.Context, objectKey string) error
}
