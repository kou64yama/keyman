package app

import (
	"flag"
	"fmt"
	"io"
)

// Executor execute the Command and handles the error.
type Executor struct {
	Stderr io.Writer
}

// Run executes the Command.
func (e *Executor) Run(cmd Command, args ...string) error {
	f := flag.NewFlagSet(cmd.Name(), flag.ContinueOnError)
	f.SetOutput(e.Stderr)
	f.Usage = Usage(cmd, f)
	cmd.Setup(f)
	if err := f.Parse(args); err != nil {
		return err
	}
	return cmd.Run(f.Args())
}

// HandleError returns the exit code for err.
func (e *Executor) HandleError(err error) int {
	if err == nil {
		return 0
	}
	if err == flag.ErrHelp {
		return 2
	}

	fmt.Fprintln(e.Stderr, err)
	return 1
}
