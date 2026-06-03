// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package cli

import (
	"fmt"
	"net"

	"github.com/charmbracelet/huh"
	"github.com/lin-snow/ech0/internal/config"
	"github.com/lin-snow/ech0/internal/di"
	tuiUtil "github.com/lin-snow/ech0/internal/util/tui"
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
		tuiUtil.PrintCLIInfo("⚠️ Start service", "Web port "+port+" is already in use; another instance may be running")
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
		tuiUtil.PrintCLIInfo("😭 Failed to start service", err.Error())
		return
	}
	if err := runtimeApp.Run(); err != nil {
		tuiUtil.PrintCLIInfo("😭 Failed to start service", err.Error())
		return
	}
	tuiUtil.PrintCLIInfo("🎉 Service stopped", "Ech0 server has stopped")
}

func DoVersion() {
	msg := fmt.Sprintf(
		"Version: v%s\nCommit: %s\nBuild Time: %s\nAuthor: %s\nWebsite: https://ech0.app/\nLicense: %s\nSource: %s\n%s",
		versionPkg.Version,
		versionPkg.Commit,
		versionPkg.BuildTime,
		versionPkg.Author,
		versionPkg.License,
		versionPkg.RepoURL,
		versionPkg.Copyright(),
	)
	if versionPkg.BuildTime == "" {
		msg = fmt.Sprintf(
			"Version: v%s\nCommit: %s\nAuthor: %s\nWebsite: https://ech0.app/\nLicense: %s\nSource: %s\n%s",
			versionPkg.Version,
			versionPkg.Commit,
			versionPkg.Author,
			versionPkg.License,
			versionPkg.RepoURL,
			versionPkg.Copyright(),
		)
	}
	item := struct{ Title, Msg string }{
		Title: "📦 Ech0",
		Msg:   msg,
	}
	tuiUtil.PrintCLIWithBox(item)
}

func DoHello() {
	tuiUtil.ClearScreen()
	tuiUtil.PrintCLIBanner()
}

func DoTui() {
	tuiUtil.ClearScreen()
	tuiUtil.PrintCLIBanner()

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
			huh.NewOption("📌 About Ech0", "version"),
			huh.NewOption("❌ Exit", "exit"),
		)

		err := huh.NewSelect[string]().
			Title("Welcome to the Ech0 TUI.").
			Options(options...).
			Value(&action).
			WithTheme(huh.ThemeCatppuccin()).
			Run()
		if err != nil {
			tuiUtil.PrintCLIInfo("😭 Operation failed", err.Error())
			return
		}

		switch action {
		case "serve":
			tuiUtil.ClearScreen()
			DoServe()
		case "servebusy":
			tuiUtil.PrintCLIInfo("ℹ️ Service status", "The web service is running in another process and cannot be stopped from here")
		case "version":
			tuiUtil.ClearScreen()
			DoVersion()
		case "exit":
			fmt.Println("👋 Thanks for using the Ech0 TUI. See you next time!")
			return
		}
	}
}
