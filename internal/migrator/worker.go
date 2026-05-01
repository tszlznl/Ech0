// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package migrator

import (
	"context"

	logUtil "github.com/lin-snow/ech0/internal/util/log"
)

type Worker struct{}

func NewWorker() *Worker {
	return &Worker{}
}

func (w *Worker) Name() string {
	return "migrator-worker"
}

func (w *Worker) Start(context.Context) error {
	logUtil.GetLogger().Info("Migrator worker started in singleton mode")
	return nil
}

func (w *Worker) Stop(context.Context) error {
	logUtil.GetLogger().Info("Migrator worker stopped")
	return nil
}
