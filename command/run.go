package command

import (
	"flag"
	"fmt"
	"io"
)

func Run(ctx *Context, args ...string) error {
	f := ctx.NewFlagSet("keyman", func(w io.Writer, f *flag.FlagSet) {
		fmt.Fprintln(w, `keyman manages secret keys.

usage: keyman [-h] [-c path] <command> [<args>]

options:`)
		f.PrintDefaults()
		fmt.Fprint(w, `
commands:
  set   save secret
  exec  execute command, arguments replaced with secret
  list  list secrets
  in    read secret from stdin and save
  out   output secret to stdout
  tee   read secret from stdin, save, and write to stdout
`)
	})

	f.StringVar(&ctx.Path, "c", ctx.Path, "specify database file")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() == 0 {
		fmt.Fprintln(ctx.Stderr, "error: no arguments")
		f.Usage()
		return flag.ErrHelp
	}

	switch f.Arg(0) {
	case "set":
		return set(ctx, f.Args()[1:]...)
	case "exec":
		return exec(ctx, f.Args()[1:]...)
	case "list":
		return list(ctx, f.Args()[1:]...)
	case "in":
		return in(ctx, f.Args()[1:]...)
	case "out":
		return out(ctx, f.Args()[1:]...)
	case "tee":
		return tee(ctx, f.Args()[1:]...)
	default:
		fmt.Fprintf(ctx.Stderr, "error: unknown command: %s\n", f.Arg(0))
		f.Usage()
		return flag.ErrHelp
	}
}
