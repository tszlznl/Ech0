package repository

import (
	"context"
	"errors"
	"time"

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

func (r *MigrationRepository) MarkCancelled(ctx context.Context, id string) error {
	now := time.Now().UTC()
	res := r.getDB(ctx).Model(&model.MigrationJob{}).
		Where("id = ? AND status IN ?", id, []string{model.MigrationStatusPending, model.MigrationStatusRunning}).
		Updates(map[string]any{
			"status":      model.MigrationStatusCancelled,
			"finished_at": &now,
			"updated_at":  now,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrMigrationJobNotFound
	}
	return nil
}

func (r *MigrationRepository) ClaimNextPendingJob(ctx context.Context) (model.MigrationJob, error) {
	tx := r.getDB(ctx).Begin()
	if tx.Error != nil {
		return model.MigrationJob{}, tx.Error
	}

	var job model.MigrationJob
	err := tx.Where("status = ?", model.MigrationStatusPending).
		Order("created_at ASC").
		First(&job).Error
	if err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.MigrationJob{}, ErrMigrationJobNotFound
		}
		return model.MigrationJob{}, err
	}

	now := time.Now().UTC()
	res := tx.Model(&model.MigrationJob{}).
		Where("id = ? AND status = ?", job.ID, model.MigrationStatusPending).
		Updates(map[string]any{
			"status":         model.MigrationStatusRunning,
			"current_phase":  model.MigrationPhaseExtracting,
			"started_at":     &now,
			"last_heartbeat": &now,
			"updated_at":     now,
		})
	if res.Error != nil {
		tx.Rollback()
		return model.MigrationJob{}, res.Error
	}
	if res.RowsAffected == 0 {
		tx.Rollback()
		return model.MigrationJob{}, ErrMigrationJobNotFound
	}

	if err := tx.Commit().Error; err != nil {
		return model.MigrationJob{}, err
	}

	job.Status = model.MigrationStatusRunning
	job.CurrentPhase = model.MigrationPhaseExtracting
	job.StartedAt = &now
	job.LastHeartbeat = &now
	return job, nil
}
