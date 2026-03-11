package bootstrap

import (
	"os"

	"github.com/lin-snow/ech0/internal/config"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"go.uber.org/zap"
)

// setEnvIfExists 尝试设置环境变量，路径不存在则忽略。
func setEnvIfExists(key, path string) {
	if st, err := os.Stat(path); err == nil && st.IsDir() {
		if err := os.Setenv(key, path); err == nil {
			logUtil.Info("set bootstrap env", zap.String("module", "bootstrap"), zap.String("key", key), zap.String("value", path))
		}
	}
}

func initHostEnv() {
	// 容错设置宿主机路径
	setEnvIfExists("HOST_PROC", "/host_proc")
	setEnvIfExists("HOST_SYS", "/host_sys")
	setEnvIfExists("HOST_ETC", "/host_etc")
	setEnvIfExists("HOST_VAR", "/host_var")
	setEnvIfExists("HOST_RUN", "/host_run")
	setEnvIfExists("HOST_ROOT", "/host_root")

	// 确保至少有默认值
	if os.Getenv("HOST_PROC") == "" {
		if err := os.Setenv("HOST_PROC", "/proc"); err != nil {
			logUtil.Warn("set HOST_PROC failed", zap.String("module", "bootstrap"), zap.Error(err))
		}
	}
	if os.Getenv("HOST_SYS") == "" {
		if err := os.Setenv("HOST_SYS", "/sys"); err != nil {
			logUtil.Warn("set HOST_SYS failed", zap.String("module", "bootstrap"), zap.Error(err))
		}
	}
	if os.Getenv("HOST_ROOT") == "" {
		if err := os.Setenv("HOST_ROOT", "/"); err != nil {
			logUtil.Warn("set HOST_ROOT failed", zap.String("module", "bootstrap"), zap.Error(err))
		}
	}
}

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
	initHostEnv()
}
