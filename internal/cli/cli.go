package cli

import (
	"flag"
	"os"
)

// Context is the CLI context: stdin, stdout and stderr.
type Context struct {
	Stdin  *os.File
	Stdout *os.File
	Stderr *os.File
}

// NewContext returns an instance of Context.
func NewContext() *Context {
	c := &Context{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	return c
}

// ExitCode returns the program exit code.
func ExitCode(err error) int {
	if err == flag.ErrHelp {
		return 2
	}
	if err != nil {
		return 1
	}
	return 0
}

// Run is entry point for CLI.
func Run(ctx *Context, args ...string) error {
	return nil
}
