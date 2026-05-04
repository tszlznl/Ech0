// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package cmd

import (
	"github.com/lin-snow/ech0/internal/cli"
	"github.com/spf13/cobra"
)

// backupCmd 是备份数据的命令
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Back up data",
	Run: func(cmd *cobra.Command, args []string) {
		cli.DoBackup()
	},
}

// init 函数用于初始化根命令和子命令
func init() {
	rootCmd.AddCommand(backupCmd)
}
