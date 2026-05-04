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
		tui.PrintCLIInfo("⚠️ Start service", "Web port "+port+" is already in use; another instance may be running")
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
		tui.PrintCLIInfo("😭 Failed to start service", err.Error())
		return
	}
	if err := runtimeApp.Run(); err != nil {
		tui.PrintCLIInfo("😭 Failed to start service", err.Error())
		return
	}
	tui.PrintCLIInfo("🎉 Service stopped", "Ech0 server has stopped")
}

func DoBackup() {
	_, backupFileName, err := backup.ExecuteBackup()
	if err != nil {
		tui.PrintCLIInfo("😭 Result", "Backup failed: "+err.Error())
		return
	}

	pwd, _ := os.Getwd()
	fullPath := filepath.Join(pwd, "data", "files", "backups", backupFileName)
	tui.PrintCLIInfo("🎉 Backup succeeded", fullPath)
}

func DoVersion() {
	item := struct{ Title, Msg string }{
		Title: "📦 Version",
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
			options = append(options, huh.NewOption("🙈 Service is running in another process", "servebusy"))
		} else {
			options = append(options, huh.NewOption("🚀 Start web service", "serve"))
		}

		options = append(options,
			huh.NewOption("🦖 Show info", "info"),
			huh.NewOption("📦 Run backup", "backup"),
			huh.NewOption("📌 Show version", "version"),
			huh.NewOption("❌ Exit", "exit"),
		)

		err := huh.NewSelect[string]().
			Title("Welcome to the Ech0 TUI.").
			Options(options...).
			Value(&action).
			WithTheme(huh.ThemeCatppuccin()).
			Run()
		if err != nil {
			tui.PrintCLIInfo("😭 Operation failed", err.Error())
			return
		}

		switch action {
		case "serve":
			tui.ClearScreen()
			DoServe()
		case "servebusy":
			tui.PrintCLIInfo("ℹ️ Service status", "The web service is running in another process and cannot be stopped from here")
		case "info":
			tui.ClearScreen()
			DoEch0Info()
		case "backup":
			DoBackup()
		case "version":
			tui.ClearScreen()
			DoVersion()
		case "exit":
			fmt.Println("👋 Thanks for using the Ech0 TUI. See you next time!")
			return
		}
	}
}
