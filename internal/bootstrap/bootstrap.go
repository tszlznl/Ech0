// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package bootstrap

import (
	"github.com/lin-snow/ech0/internal/config"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
)

func initLogger() {
	cfg := config.Config()
	logUtil.InitLoggerWithConfig(logUtil.LogConfig{
		Level:   cfg.Log.Level,
		Format:  cfg.Log.Format,
		Console: cfg.Log.Console,
		File: logUtil.FileConfig{
			Enable:     cfg.Log.FileEnable,
			Filename:   cfg.Log.FilePath,
			MaxSize:    cfg.Log.FileMaxSize,
			MaxBackups: cfg.Log.FileMaxBackups,
			MaxAge:     cfg.Log.FileMaxAge,
			Compress:   cfg.Log.FileCompress,
		},
		Stream: logUtil.StreamConfig{
			BufferSize:      cfg.Log.BufferSize,
			RecentSize:      cfg.Log.RecentSize,
			DropPolicy:      cfg.Log.DropPolicy,
			FlushBatch:      cfg.Log.FlushBatch,
			FlushIntervalMs: cfg.Log.FlushIntervalMs,
		},
	})
}

func initConfig() {
	config.Config()
}

// Bootstrap 执行应用启动阶段所需的基础初始化流程。
func Bootstrap() {
	initConfig()
	initLogger()
}
