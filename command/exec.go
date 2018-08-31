package command

import (
	"flag"
	"io"

	"github.com/boltdb/bolt"
	"github.com/kou64yama/keyman"
)

func exec(ctx *Context, args ...string) error {
	f := ctx.NewFlagSet("keyman exec", func(w io.Writer, f *flag.FlagSet) {
	})
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() < 1 {
		f.Usage()
		return flag.ErrHelp
	}

	db, err := bolt.Open(ctx.Path, 0600, nil)
	if err != nil {
		return err
	}

	return db.View(func(tx *bolt.Tx) error {
		k := keyman.New(tx)
		cmd, err := k.Command(f.Args()[0], f.Args()[1:]...)
		if err != nil {
			return err
		}

		cmd.Stdin = ctx.Stdin
		cmd.Stdout = ctx.Stdout
		cmd.Stderr = ctx.Stderr
		return cmd.Run()
	})
}
