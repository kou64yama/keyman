package command

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"

	"github.com/boltdb/bolt"
	"github.com/kou64yama/keyman"
)

func list(ctx *Context, args ...string) error {
	var verbose bool
	f := ctx.NewFlagSet("keyman list", func(w io.Writer, f *flag.FlagSet) {
	})
	f.BoolVar(&verbose, "v", false, "verbose output")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 {
		flag.Usage()
		return flag.ErrHelp
	}

	db, err := bolt.Open(ctx.Path, 0600, nil)
	if err != nil {
		return err
	}

	return db.View(func(tx *bolt.Tx) error {
		k := keyman.New(tx)
		return k.ForEach(func(name string, rev []byte, x keyman.Context) error {
			fmt.Fprintln(ctx.Stdout, name)

			if verbose {
				h, err := x.OpenHistoryBucket(name)
				if err != nil {
					return err
				}

				m, err := h.Get(rev)
				if err != nil {
					return err
				}

				fmt.Fprintln(ctx.Stdout, "    Revision:", binary.BigEndian.Uint64(rev))
				fmt.Fprintln(ctx.Stdout, "    Created At:", m.CreatedAt.Local())
				fmt.Fprintln(ctx.Stdout)
			}

			return nil
		})
	})
}
