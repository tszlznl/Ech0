package cli

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/lin-snow/ech0/internal/app"
	"github.com/lin-snow/ech0/internal/backup"
	"github.com/lin-snow/ech0/internal/config"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	"github.com/lin-snow/ech0/internal/tui"
)

var runtimeApp *app.App

// SetApp 注入应用实例。
func SetApp(a *app.App) {
	runtimeApp = a
}

// isWebPortInUse 检查 Web 端口是否已被占用（通常表示已有实例在运行）
func isWebPortInUse() bool {
	port := config.Config().Server.Port
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return true
	}
	_ = ln.Close()
	return false
}

// canStartWebServer 检查当前进程或系统端口是否允许启动 Web 服务
func canStartWebServer() bool {
	if runtimeApp == nil {
		tui.PrintCLIInfo("⚠️ 启动服务", "应用未初始化")
		return false
	}

	if runtimeApp.IsWebRunning() {
		tui.PrintCLIInfo("⚠️ 启动服务", "Web 服务已在当前进程中运行")
		return false
	}

	if isWebPortInUse() {
		port := config.Config().Server.Port
		tui.PrintCLIInfo("⚠️ 启动服务", "Web 端口 "+port+" 已被占用，可能已有实例在运行")
		return false
	}

	return true
}

// DoServe 启动服务
func DoServe() {
	if !canStartWebServer() {
		return
	}

	if err := runtimeApp.StartWeb(context.Background()); err != nil {
		tui.PrintCLIInfo("😭 启动服务失败", err.Error())
	}
}

// DoServeWithBlock 阻塞当前线程，直到服务器停止
func DoServeWithBlock() {
	if !canStartWebServer() {
		return
	}

	if err := runtimeApp.StartWeb(context.Background()); err != nil {
		tui.PrintCLIInfo("😭 启动服务失败", err.Error())
		return
	}

	// 阻塞主线程，直到接收到终止信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 创建 context，最大等待 5 秒优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := runtimeApp.StopWeb(ctx); err != nil {
		tui.PrintCLIInfo("❌ 服务停止", "服务器强制关闭")
		os.Exit(1)
	}
	tui.PrintCLIInfo("🎉 停止服务成功", "Ech0 服务器已停止")
}

// DoStopServe 停止服务
func DoStopServe() {
	if runtimeApp == nil || !runtimeApp.IsWebRunning() {
		tui.PrintCLIInfo("⚠️ 停止服务", "Ech0 服务器未启动")
		return
	}

	// 创建 context，最大等待 5 秒优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := runtimeApp.StopWeb(ctx); err != nil {
		tui.PrintCLIInfo("😭 停止服务失败", err.Error())
		return
	}

	tui.PrintCLIInfo("🎉 停止服务成功", "Ech0 服务器已停止")
}

// DoBackup 执行备份
func DoBackup() {
	_, backupFileName, err := backup.ExecuteBackup()
	if err != nil {
		// 处理错误
		tui.PrintCLIInfo("😭 执行结果", "备份失败: "+err.Error())
		return
	}

	// 获取PWD环境变量
	pwd, _ := os.Getwd()
	fullPath := filepath.Join(pwd, "backup", backupFileName)

	tui.PrintCLIInfo("🎉 备份成功", fullPath)
}

// DoRestore 执行恢复
func DoRestore(backupFilePath string) {
	err := backup.ExecuteRestore(backupFilePath)
	if err != nil {
		// 处理错误
		tui.PrintCLIInfo("😭 执行结果", "恢复失败: "+err.Error())
		return
	}
	tui.PrintCLIInfo("🎉 恢复成功", "已从备份文件 "+backupFilePath+" 中恢复数据")
}

// DoVersion 打印版本信息
func DoVersion() {
	item := struct{ Title, Msg string }{
		Title: "📦 当前版本",
		Msg:   "v" + commonModel.Version,
	}
	tui.PrintCLIWithBox(item)
}

// DoEch0Info 打印 Ech0 信息
func DoEch0Info() {
	if _, err := fmt.Fprintln(os.Stdout, tui.GetEch0Info()); err != nil {
		fmt.Fprintf(os.Stderr, "failed to print ech0 info: %v\n", err)
	}
}

// DoHello 打印 Ech0 Logo
func DoHello() {
	tui.ClearScreen()
	tui.PrintCLIBanner()
}

// DoTui 执行 TUI
func DoTui() {
	// 清除屏幕当前字符
	tui.ClearScreen()
	// 打印 ASCII 风格 Banner
	tui.PrintCLIBanner()

	for {
		// 换行
		fmt.Println()

		var action string
		var options []huh.Option[string]

		if runtimeApp != nil && runtimeApp.IsWebRunning() {
			options = append(options, huh.NewOption("🛑 停止 Web 服务", "stopserve"))
		} else if isWebPortInUse() {
			options = append(options, huh.NewOption("🙈 服务已在其他进程中运行", "servebusy"))
		} else {
			options = append(options, huh.NewOption("🚀 启动 Web 服务", "serve"))
		}

		options = append(options,
			huh.NewOption("🦖 查看信息", "info"),
			huh.NewOption("📦 执行备份", "backup"),
			huh.NewOption("💾 恢复备份", "restore"),
			huh.NewOption("📌 查看版本", "version"),
			huh.NewOption("❌ 退出", "exit"),
		)

		err := huh.NewSelect[string]().
			Title("欢迎使用 Ech0 TUI .").
			Options(options...).
			Value(&action).
			WithTheme(huh.ThemeCatppuccin()).
			Run()
		if err != nil {
			log.Fatal(err)
		}

		switch action {
		case "serve":
			tui.ClearScreen()
			DoServe()
		case "servebusy":
			tui.PrintCLIInfo("ℹ️ Web 服务状态", "当前 Web 服务由其他进程运行，无法在此进程内停止")
		case "stopserve":
			tui.ClearScreen()
			DoStopServe()
		case "info":
			tui.ClearScreen()
			DoEch0Info()
		case "backup":
			DoBackup()
		case "restore":
			// 如果服务器已经启动，则先停止服务器
			if runtimeApp != nil && runtimeApp.IsWebRunning() {
				tui.PrintCLIInfo("⚠️ 警告", "恢复数据前请先停止服务器")
			} else {
				// 获取备份文件路径
				var path string
				_ = huh.NewInput().
					Title("请输入备份文件路径").
					Value(&path).
					Run()
				path = strings.TrimSpace(path)
				if path != "" {
					DoRestore(path)
				} else {
					tui.PrintCLIInfo("⚠️ 跳过", "未输入备份路径")
				}
			}
		case "version":
			tui.ClearScreen()
			DoVersion()
		case "exit":
			fmt.Println("👋 感谢使用 Ech0 TUI，期待下次再见")
			return
		}
	}
}
