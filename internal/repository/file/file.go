package repository

import (
	"context"
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
