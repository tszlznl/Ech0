// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package cmd

import (
	"github.com/lin-snow/ech0/internal/cli"
	"github.com/spf13/cobra"
)

// serveCmd 是启动 Web 服务的命令
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the web server (blocking)",
	Run: func(cmd *cobra.Command, args []string) {
		cli.DoServeWithBlock()
	},
}

// webCmd 是仅启动 Web 服务并阻塞的命令
var webCmd = &cobra.Command{
	Use:        "web",
	Short:      "(Compatibility alias) Start the web server (blocking)",
	Deprecated: "use `serve` instead; `web` will be removed in a future release",
	Hidden:     true,
	Run: func(cmd *cobra.Command, args []string) {
		cli.DoServeWithBlock()
	},
}

// init 函数用于初始化根命令和子命令
func init() {
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(webCmd)
}
