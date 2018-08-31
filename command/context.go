package command

import (
	"flag"
	"io"
)

type Context struct {
	Stdin        io.Reader
	Stdout       io.Writer
	Stderr       io.Writer
	Path         string
	ReadPassword func() ([]byte, error)
}

func (c *Context) NewFlagSet(name string, usage func(writer io.Writer, f *flag.FlagSet)) *flag.FlagSet {
	f := flag.NewFlagSet(name, flag.ContinueOnError)
	f.Usage = func() {
		usage(c.Stderr, f)
	}
	f.SetOutput(c.Stderr)
	return f
}
