package main

import (
	"github.com/caarlos0/timer/cmd"
	"os"
)

func main() {
	if err := cmd.Root.Execute(); err != nil {
		os.Exit(1)
	}
}
