package migrator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/lin-snow/ech0/internal/config"
	migrationModel "github.com/lin-snow/ech0/internal/model/migration"
	migrationRepository "github.com/lin-snow/ech0/internal/repository/migration"
	migratorService "github.com/lin-snow/ech0/internal/service/migrator"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"go.uber.org/zap"
)

type Worker struct {
	service migratorService.Service
	enabled bool

	pollInterval time.Duration
	batchSize    int
	rateLimit    int

	started bool
	stopCh  chan struct{}
	wg      sync.WaitGroup
	mu      sync.Mutex
}

func NewWorker(service migratorService.Service) *Worker {
	cfg := config.Config().Migration
	return &Worker{
		service:      service,
		enabled:      cfg.WorkerEnabled,
		pollInterval: 2 * time.Second,
		batchSize:    cfg.BatchSize,
		rateLimit:    cfg.RateLimitPerSec,
		stopCh:       make(chan struct{}),
	}
}

func (w *Worker) Name() string {
	return "migrator-worker"
}

func (w *Worker) Start(context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.enabled || w.started {
		return nil
	}
	w.wg.Add(1)
	go w.loop()
	w.started = true
	logUtil.GetLogger().Info("Migrator worker started")
	return nil
}

func (w *Worker) Stop(context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.started {
		return nil
	}
	close(w.stopCh)
	w.wg.Wait()
	w.started = false
	return nil
}

func (w *Worker) loop() {
	defer w.wg.Done()
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopCh:
			return
		default:
		}

		job, err := w.service.ClaimNextPendingJob(context.Background())
		if err != nil {
			if errors.Is(err, migrationRepository.ErrMigrationJobNotFound) {
				select {
				case <-ticker.C:
					continue
				case <-w.stopCh:
					return
				}
			}
			logUtil.GetLogger().Error("claim migration job failed", zap.Error(err))
			select {
			case <-ticker.C:
				continue
			case <-w.stopCh:
				return
			}
		}

		if runErr := w.runJob(job); runErr != nil {
			logUtil.GetLogger().Error("run migration job failed",
				zap.String("jobID", job.ID),
				zap.Error(runErr),
			)
		}
	}
}

func (w *Worker) runJob(job migrationModel.MigrationJob) error {
	runner, err := BuildRunner(job.SourceType, job.CreatedBy)
	if err != nil {
		return w.markJobFailed(job, fmt.Sprintf("build runner failed: %v", err))
	}

	sourcePayload := map[string]any{}
	if len(job.SourcePayload) > 0 {
		if err := json.Unmarshal(job.SourcePayload, &sourcePayload); err != nil {
			return w.markJobFailed(job, fmt.Sprintf("unmarshal source payload failed: %v", err))
		}
	}

	report := map[string]any{
		"job_id":       job.ID,
		"source_type":  job.SourceType,
		"source_ver":   job.SourceVersion,
		"started_at":   time.Now().UTC().Format(time.RFC3339),
		"failed_items": []FailedItem{},
	}

	for {
		current, err := w.service.GetJobModel(context.Background(), job.ID)
		if err != nil {
			return err
		}
		if current.Status == migrationModel.MigrationStatusCancelled {
			return nil
		}

		current.CurrentPhase = migrationModel.MigrationPhaseExtracting
		now := time.Now().UTC()
		current.LastHeartbeat = &now
		if err := w.service.UpdateJob(context.Background(), &current); err != nil {
			return err
		}

		outcome, err := runner.RunBatch(context.Background(), ExtractRequest{
			SourcePayload: sourcePayload,
			Checkpoint:    current.Checkpoint,
			BatchSize:     w.batchSize,
		})
		if err != nil {
			return w.markJobFailed(current, err.Error())
		}

		current.CurrentPhase = migrationModel.MigrationPhaseLoading
		current.Total = maxInt64(current.Total, outcome.TotalHint)
		current.Checkpoint = outcome.NextCheckpoint
		current.Processed += int64(len(outcome.Failed)) + outcome.Loaded
		current.SuccessCount += outcome.Loaded
		current.FailCount += int64(len(outcome.Failed))

		if len(outcome.Failed) > 0 {
			current.ErrorSummary = fmt.Sprintf("当前累计失败记录: %d", current.FailCount)
			failedJSON, marshalErr := json.Marshal(outcome.Failed)
			if marshalErr == nil {
				current.FailedItems = failedJSON
			}
		}

		if !outcome.HasMore {
			current.CurrentPhase = migrationModel.MigrationPhaseCompleted
			current.Status = migrationModel.MigrationStatusSuccess
			finished := time.Now().UTC()
			current.FinishedAt = &finished
			report["finished_at"] = finished.Format(time.RFC3339)
			report["success"] = current.SuccessCount
			report["failed"] = current.FailCount
			report["status"] = current.Status
			if repJSON, marshalErr := json.Marshal(report); marshalErr == nil {
				current.Report = repJSON
			}
		}

		if err := w.service.UpdateJob(context.Background(), &current); err != nil {
			return err
		}

		if !outcome.HasMore {
			return nil
		}

		if w.rateLimit > 0 {
			time.Sleep(time.Second / time.Duration(w.rateLimit))
		}
	}
}

func (w *Worker) markJobFailed(job migrationModel.MigrationJob, reason string) error {
	job.Status = migrationModel.MigrationStatusFailed
	job.FatalError = reason
	job.CurrentPhase = migrationModel.MigrationPhaseReporting
	finished := time.Now().UTC()
	job.FinishedAt = &finished
	return w.service.UpdateJob(context.Background(), &job)
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
