package keyman

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

type Context interface {
}

type Command interface {
	Name() string
	Help() string
	Context(func(ctx Context) Context)
	Flag(func(ctx Context, f *flag.FlagSet))
	Handle(func(ctx Context, args ...string) error)
	SubCommand(synopsis string, help string) Command
	Run(ctx Context, args ...string) error
}

type cmd struct {
	Command
	name     string
	synopsis string
	help     string
	context  func(ctx Context) Context
	flagSet  func(ctx Context, f *flag.FlagSet)
	handle   func(ctx Context, args ...string) error
	subMap   map[string]Command
	subs     []Command
	stdin    io.Reader
	stdout   io.Writer
	stderr   io.Writer
}

func NewCommand(synopsis string, help string) Command {
	return &cmd{
		name:     strings.Split(synopsis, " ")[0],
		synopsis: synopsis,
		help:     help,
		subMap:   make(map[string]Command),
		stdin:    os.Stdin,
		stdout:   os.Stdout,
		stderr:   os.Stderr,
	}
}

func (c *cmd) Name() string { return c.name }
func (c *cmd) Help() string { return c.help }

func (c *cmd) Context(context func(ctx Context) Context) {
	c.context = context
}

func (c *cmd) Flag(flagSet func(ctx Context, f *flag.FlagSet)) {
	c.flagSet = flagSet
}

func (c *cmd) Handle(handle func(ctx Context, args ...string) error) {
	c.handle = handle
}

func (c *cmd) SubCommand(synopsis string, help string) Command {
	sub := &cmd{
		name:     strings.Split(synopsis, " ")[0],
		synopsis: c.name + " " + synopsis,
		help:     help,
		subMap:   make(map[string]Command),
		stdin:    c.stdin,
		stdout:   c.stdout,
		stderr:   c.stderr,
	}
	c.subMap[sub.Name()] = sub
	c.subs = append(c.subs, sub)
	return sub
}

func (c *cmd) Run(ctx Context, args ...string) error {
	if c.context != nil {
		ctx = c.context(ctx)
	}

	f := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	f.SetOutput(c.stderr)
	f.Usage = func() {
		fmt.Fprintln(c.stderr, "Usage:", c.synopsis)
		if len(c.help) > 0 {
			fmt.Fprintf(c.stderr, "\n%s\n", c.help)
		}

		if c.flagSet != nil {
			fmt.Fprintln(c.stderr, "\nOptions:")
			f.PrintDefaults()
		}

		if len(c.subs) > 0 {
			fmt.Fprintln(c.stderr, "\nActions:")
			for _, sub := range c.subs {
				fmt.Fprintf(c.stderr, "  %s\t%s\n", sub.Name(), sub.Help())
			}
		}
	}
	if c.flagSet != nil {
		c.flagSet(ctx, f)
	}

	if err := f.Parse(args); err != nil {
		return flag.ErrHelp
	}

	if f.NArg() > 0 {
		if sub := c.subMap[f.Arg(0)]; sub != nil {
			return sub.Run(ctx, f.Args()[1:]...)
		}
	}
	if c.handle != nil {
		return c.handle(ctx, f.Args()...)
	}

	if len(args) == 0 {
		fmt.Fprintln(c.stderr, "no arguments")
	} else {
		fmt.Fprintln(c.stderr, "invalid arguments:", strings.Join(args, " "))
	}
	f.Usage()
	return flag.ErrHelp
}
