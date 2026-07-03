// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package scheduled

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-co-op/gocron/v2"
	fileService "github.com/lin-snow/ech0/internal/service/file"
	logUtil "github.com/lin-snow/ech0/pkg/log"
)

// Cleanup 周期清理过期的临时/孤儿文件。
type Cleanup struct {
	fileService fileService.Service
}

func NewCleanup(fileSvc fileService.Service) *Cleanup {
	return &Cleanup{fileService: fileSvc}
}

func (c *Cleanup) Name() string { return "cleanup-temp-files" }

// Schedule 每三天清理一次孤儿文件。
func (c *Cleanup) Schedule(_ context.Context, s gocron.Scheduler) error {
	_, err := s.NewJob(
		gocron.DurationJob(72*time.Hour),
		gocron.NewTask(func() {
			if err := c.fileService.CleanupOrphanFiles(); err != nil {
				logUtil.GetLogger().Error("Failed to clean up temporary files",
					slog.String("module", logModule), logUtil.Err(err))
			}
		}),
	)
	if err != nil {
		logUtil.GetLogger().Error("Failed to schedule cleanup task",
			slog.String("module", logModule), logUtil.Err(err))
	}
	return err
}
