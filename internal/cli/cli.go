// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package cli

import (
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/lin-snow/ech0/internal/backup"
	"github.com/lin-snow/ech0/internal/config"
	"github.com/lin-snow/ech0/internal/di"
	"github.com/lin-snow/ech0/internal/tui"
	versionPkg "github.com/lin-snow/ech0/internal/version"
)

func isWebPortInUse() bool {
	port := config.Config().Server.Port
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return true
	}
	_ = ln.Close()
	return false
}

func canStartWebServer() bool {
	if isWebPortInUse() {
		port := config.Config().Server.Port
		tui.PrintCLIInfo("⚠️ 启动服务", "Web 端口 "+port+" 已被占用，可能已有实例在运行")
		return false
	}
	return true
}

func DoServe() {
	DoServeWithBlock()
}

func DoServeWithBlock() {
	if !canStartWebServer() {
		return
	}
	runtimeApp, err := di.BuildApp()
	if err != nil {
		tui.PrintCLIInfo("😭 启动服务失败", err.Error())
		return
	}
	if err := runtimeApp.Run(); err != nil {
		tui.PrintCLIInfo("😭 启动服务失败", err.Error())
		return
	}
	tui.PrintCLIInfo("🎉 停止服务成功", "Ech0 服务器已停止")
}

func DoBackup() {
	_, backupFileName, err := backup.ExecuteBackup()
	if err != nil {
		tui.PrintCLIInfo("😭 执行结果", "备份失败: "+err.Error())
		return
	}

	pwd, _ := os.Getwd()
	fullPath := filepath.Join(pwd, "data", "files", "backups", backupFileName)
	tui.PrintCLIInfo("🎉 备份成功", fullPath)
}

func DoVersion() {
	item := struct{ Title, Msg string }{
		Title: "📦 当前版本",
		Msg:   "v" + versionPkg.Version,
	}
	tui.PrintCLIWithBox(item)
}

func DoEch0Info() {
	if _, err := fmt.Fprintln(os.Stdout, tui.GetEch0Info()); err != nil {
		fmt.Fprintf(os.Stderr, "failed to print ech0 info: %v\n", err)
	}
}

func DoHello() {
	tui.ClearScreen()
	tui.PrintCLIBanner()
}

func DoTui() {
	tui.ClearScreen()
	tui.PrintCLIBanner()

	for {
		fmt.Println()

		var action string
		var options []huh.Option[string]

		if isWebPortInUse() {
			options = append(options, huh.NewOption("🙈 服务已在其他进程中运行", "servebusy"))
		} else {
			options = append(options, huh.NewOption("🚀 启动 Web 服务", "serve"))
		}

		options = append(options,
			huh.NewOption("🦖 查看信息", "info"),
			huh.NewOption("📦 执行备份", "backup"),
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
			tui.PrintCLIInfo("😭 操作失败", err.Error())
			return
		}

		switch action {
		case "serve":
			tui.ClearScreen()
			DoServe()
		case "servebusy":
			tui.PrintCLIInfo("ℹ️ Web 服务状态", "当前 Web 服务由其他进程运行，无法在此进程内停止")
		case "info":
			tui.ClearScreen()
			DoEch0Info()
		case "backup":
			DoBackup()
		case "version":
			tui.ClearScreen()
			DoVersion()
		case "exit":
			fmt.Println("👋 感谢使用 Ech0 TUI，期待下次再见")
			return
		}
	}
}
