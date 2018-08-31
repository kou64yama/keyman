package command

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"

	"github.com/boltdb/bolt"
	"github.com/kou64yama/keyman"
)

func set(ctx *Context, args ...string) error {
	f := ctx.NewFlagSet("keyman set", func(w io.Writer, f *flag.FlagSet) {
	})
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 1 {
		f.Usage()
		return flag.ErrHelp
	}

	fmt.Fprint(ctx.Stderr, "Password: ")
	v, err := ctx.ReadPassword()
	fmt.Fprintln(ctx.Stderr)
	if err != nil {
		return err
	}

	db, err := bolt.Open(ctx.Path, 0600, nil)
	if err != nil {
		return err
	}

	return db.Update(func(tx *bolt.Tx) error {
		k := keyman.New(tx)
		rev, err := k.Put(f.Arg(0), v)
		if err != nil {
			return err
		}

		fmt.Fprintf(ctx.Stdout, "revision %d\n", binary.BigEndian.Uint64(rev))
		return nil
	})
}
