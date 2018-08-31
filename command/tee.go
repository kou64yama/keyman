package command

import (
	"flag"
	"io"
	"io/ioutil"

	"github.com/boltdb/bolt"
	"github.com/kou64yama/keyman"
)

func tee(ctx *Context, args ...string) error {
	f := ctx.NewFlagSet("keyman tee", func(w io.Writer, f *flag.FlagSet) {
	})
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 1 {
		f.Usage()
		return flag.ErrHelp
	}

	v, err := ioutil.ReadAll(ctx.Stdin)
	if err != nil {
		return err
	}

	db, err := bolt.Open(ctx.Path, 0600, nil)
	if err != nil {
		return err
	}

	return db.Update(func(tx *bolt.Tx) error {
		k := keyman.New(tx)
		if _, err := k.Put(f.Arg(0), v); err != nil {
			return err
		}

		ctx.Stdout.Write(v)
		return nil
	})
}
