package cmd

import (
	"flag"
	"fmt"
	"io"
	"keyman"
	"keyman/internal/app"
	"os"
)

// MainCommand is the main command.
type MainCommand struct {
	Usage   func()
	Stdout  io.Writer
	Stderr  io.Writer
	Version bool
	Actions []app.Command
}

// Name returns os.Args[0].
func (c *MainCommand) Name() string {
	return os.Args[0]
}

// Synopsis returns the synopsis of the MainCommand.
func (c *MainCommand) Synopsis() []string {
	s := []string{os.Args[0] + " [options]"}
	for _, act := range c.Actions {
		s = append(s, act.Synopsis()...)
	}
	return s
}

// Setup sets up MainCommand and flag.FlagSet.
func (c *MainCommand) Setup(f *flag.FlagSet) {
	f.BoolVar(&c.Version, "V", false, "show version")
	c.Usage = f.Usage
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Actions = []app.Command{}
}

// Run executes the MainCommand.
func (c *MainCommand) Run(args []string) error {
	if c.Version {
		fmt.Fprintf(c.Stdout, "Keyman %s\n", keyman.Version())
		return nil
	}

	if len(args) == 0 {
		c.Usage()
		return flag.ErrHelp
	}

	for _, cmd := range c.Actions {
		if cmd.Name() == args[0] {
			exec := app.Executor{Stderr: c.Stderr}
			return exec.Run(cmd, args[1:]...)
		}
	}
	fmt.Fprintf(c.Stderr, "action provided but not defined: %s\n", args[0])
	return flag.ErrHelp
}
