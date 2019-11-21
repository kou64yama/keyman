package main

import (
	"os"

	"github.com/kou64yama/keyman/internal/cli"
)

func main() {
	ctx := cli.NewContext()
	err := cli.Run(ctx, os.Args[1:]...)
	os.Exit(cli.ExitCode(err))
}
