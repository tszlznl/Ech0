package repository

import (
	"context"
	"strings"
	"time"

	model "github.com/lin-snow/ech0/internal/model/file"
	fileService "github.com/lin-snow/ech0/internal/service/file"
	"github.com/lin-snow/ech0/internal/transaction"
	"gorm.io/gorm"
)

type FileRepository struct {
	db func() *gorm.DB
}

var _ fileService.FileRepository = (*FileRepository)(nil)

func NewFileRepository(dbProvider func() *gorm.DB) *FileRepository {
	return &FileRepository{db: dbProvider}
}

func (r *FileRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := transaction.TxFromContext(ctx); ok {
		return tx
	}
	return r.db()
}

func (r *FileRepository) Create(ctx context.Context, file *model.File) error {
	return r.getDB(ctx).Create(file).Error
}

func (r *FileRepository) GetByID(ctx context.Context, id string) (*model.File, error) {
	var f model.File
	if err := r.getDB(ctx).Where("id = ?", id).First(&f).Error; err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *FileRepository) GetByKey(ctx context.Context, key string) (*model.File, error) {
	var f model.File
	if err := r.getDB(ctx).Where("key = ?", key).First(&f).Error; err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *FileRepository) GetByRoute(
	ctx context.Context,
	storageType, provider, bucket, key string,
) (*model.File, error) {
	var f model.File
	if err := r.getDB(ctx).
		Where("storage_type = ? AND provider = ? AND bucket = ? AND key = ?",
			storageType, provider, bucket, key).
		First(&f).Error; err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *FileRepository) ListByStorageTypeAndSearch(
	ctx context.Context,
	storageType string,
	search string,
	page int,
	pageSize int,
) ([]model.File, int64, error) {
	db := r.getDB(ctx).Model(&model.File{})
	if storageType != "" {
		db = db.Where("storage_type = ?", storageType)
	}
	if trimmed := strings.TrimSpace(search); trimmed != "" {
		like := "%" + trimmed + "%"
		db = db.Where("name LIKE ? OR key LIKE ?", like, like)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []model.File{}, 0, nil
	}

	offset := (page - 1) * pageSize
	var files []model.File
	err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&files).Error
	return files, total, err
}

func (r *FileRepository) ListByStorageTypeAndURLs(
	ctx context.Context,
	storageType string,
	urls []string,
) ([]model.File, error) {
	if len(urls) == 0 {
		return []model.File{}, nil
	}
	var files []model.File
	err := r.getDB(ctx).
		Where("storage_type = ? AND url IN ?", storageType, urls).
		Find(&files).Error
	return files, err
}

func (r *FileRepository) UpdateMetaByID(
	ctx context.Context,
	id string,
	size int64,
	width *int,
	height *int,
	contentType *string,
) (*model.File, error) {
	updates := map[string]any{
		"size": size,
	}
	if width != nil {
		updates["width"] = *width
	}
	if height != nil {
		updates["height"] = *height
	}
	if contentType != nil {
		updates["content_type"] = *contentType
	}

	if err := r.getDB(ctx).Model(&model.File{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *FileRepository) Delete(ctx context.Context, id string) error {
	return r.getDB(ctx).Where("id = ?", id).Delete(&model.File{}).Error
}

func (r *FileRepository) DeleteByRoute(
	ctx context.Context,
	storageType, provider, bucket, key string,
) error {
	return r.getDB(ctx).
		Where("storage_type = ? AND provider = ? AND bucket = ? AND key = ?",
			storageType, provider, bucket, key).
		Delete(&model.File{}).Error
}

func (r *FileRepository) GetOrphanFiles(ctx context.Context, olderThan time.Time) ([]model.File, error) {
	var files []model.File
	err := r.getDB(ctx).
		Where("created_at < ?", olderThan).
		Where("id NOT IN (?)",
			r.getDB(ctx).Table("echo_files").Select("file_id"),
		).
		Find(&files).Error
	return files, err
}

func (r *FileRepository) GetByCategory(ctx context.Context, category string) ([]model.File, error) {
	var files []model.File
	err := r.getDB(ctx).Where("category = ?", category).Find(&files).Error
	return files, err
}
