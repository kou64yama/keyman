//+build !test

package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"path"
	"syscall"

	"github.com/kou64yama/keyman/command"
	"golang.org/x/crypto/ssh/terminal"
)

func main() {
	u, err := user.Current()
	if err != nil {
		panic(err)
	}

	ctx := &command.Context{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Path:   path.Join(u.HomeDir, ".keyman.db"),
		ReadPassword: func() ([]byte, error) {
			return terminal.ReadPassword(int(syscall.Stdin))
		},
	}

	exit(ctx, command.Run(ctx, os.Args[1:]...))
}

func exit(ctx *command.Context, err error) {
	if err == nil {
		os.Exit(0)
	}
	if err == flag.ErrHelp {
		os.Exit(2)
	}

	fmt.Fprintln(ctx.Stderr, err)
	os.Exit(1)
}
