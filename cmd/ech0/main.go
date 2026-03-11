package main

import (
	"github.com/lin-snow/ech0/cmd"
	"github.com/lin-snow/ech0/internal/bootstrap"
)

func main() {
	bootstrap.Bootstrap()
	cmd.Execute()
}
