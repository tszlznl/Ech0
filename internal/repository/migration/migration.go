package repository

import (
	"context"
	"errors"

	model "github.com/lin-snow/ech0/internal/model/migration"
	"github.com/lin-snow/ech0/internal/transaction"
	"gorm.io/gorm"
)

var ErrMigrationJobNotFound = errors.New("migration job not found")

type MigrationRepository struct {
	db func() *gorm.DB
}

func NewMigrationRepository(db func() *gorm.DB) *MigrationRepository {
	return &MigrationRepository{db: db}
}

func (r *MigrationRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := transaction.TxFromContext(ctx); ok {
		return tx
	}
	return r.db()
}

func (r *MigrationRepository) CreateJob(ctx context.Context, job *model.MigrationJob) error {
	return r.getDB(ctx).Create(job).Error
}

func (r *MigrationRepository) GetJobByID(ctx context.Context, id string) (model.MigrationJob, error) {
	var job model.MigrationJob
	err := r.getDB(ctx).Where("id = ?", id).First(&job).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.MigrationJob{}, ErrMigrationJobNotFound
	}
	return job, err
}

func (r *MigrationRepository) UpdateJob(ctx context.Context, job *model.MigrationJob) error {
	return r.getDB(ctx).Save(job).Error
}
