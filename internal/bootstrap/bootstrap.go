package bootstrap

import (
	"fmt"
	"os"

	"github.com/lin-snow/ech0/internal/config"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
)

// setEnvIfExists 尝试设置环境变量，路径不存在则忽略。
func setEnvIfExists(key, path string) {
	if st, err := os.Stat(path); err == nil && st.IsDir() {
		if err := os.Setenv(key, path); err == nil {
			fmt.Printf("[bootstrap] %s=%s\n", key, path)
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
			fmt.Printf("[bootstrap] failed to set HOST_PROC: %v\n", err)
		}
	}
	if os.Getenv("HOST_SYS") == "" {
		if err := os.Setenv("HOST_SYS", "/sys"); err != nil {
			fmt.Printf("[bootstrap] failed to set HOST_SYS: %v\n", err)
		}
	}
	if os.Getenv("HOST_ROOT") == "" {
		if err := os.Setenv("HOST_ROOT", "/"); err != nil {
			fmt.Printf("[bootstrap] failed to set HOST_ROOT: %v\n", err)
		}
	}
}

func initLogger() {
	logUtil.InitLogger()
}

func initConfig() {
	config.Config()
}

// Bootstrap 执行应用启动阶段所需的基础初始化流程。
func Bootstrap() {
	initHostEnv()
	initLogger()
	initConfig()
}
