package main

import (
	"github.com/lin-snow/ech0/cmd"
	"github.com/lin-snow/ech0/internal/bootstrap"
	_ "time/tzdata"
)

func main() {
	bootstrap.Bootstrap()
	cmd.Execute()
}
