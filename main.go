package main

import (
	"github.com/lin-snow/ech0/cmd"
	_ "github.com/lin-snow/ech0/internal/bootstrap"
	"github.com/lin-snow/ech0/internal/config"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
)

func init() {
	// Logger
	logUtil.InitLogger()

	// Config
	config.Config()
}

func main() {
	cmd.Execute()
}
