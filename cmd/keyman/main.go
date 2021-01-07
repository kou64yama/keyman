package main

import (
	"keyman/internal/app"
	"keyman/internal/cmd"
	"os"
)

func main() {
	args := os.Args[1:]
	code := app.Run(new(cmd.MainCommand), args...)
	os.Exit(code)
}
