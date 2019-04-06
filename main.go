package main

import (
	"os"

	"github.com/karadalex/fabricbeat/cmd"

	_ "github.com/karadalex/fabricbeat/include"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
