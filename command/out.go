package command

import (
	"flag"
	"io"

	"github.com/boltdb/bolt"
	"github.com/kou64yama/keyman"
)

func out(ctx *Context, args ...string) error {
	f := ctx.NewFlagSet("keyman out", func(w io.Writer, f *flag.FlagSet) {
	})
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 1 {
		f.Usage()
		return flag.ErrHelp
	}

	db, err := bolt.Open(ctx.Path, 0600, nil)
	if err != nil {
		return err
	}

	return db.View(func(tx *bolt.Tx) error {
		k := keyman.New(tx)
		v, err := k.Get(f.Arg(0))
		if err != nil {
			return err
		}

		ctx.Stdout.Write(v)
		return nil
	})
}
