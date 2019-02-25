package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/kou64yama/keyman"
)

type appCtx struct {
	keyman.Context
	path string
}

type setCtx struct {
	*appCtx
	rev uint64
}

type listCtx struct {
	*appCtx
	long bool
}

type logCtx struct {
	*appCtx
	limit uint64
}

func main() {
	err := run(os.Args...)
	exit(err)
}

func exit(err error) {
	if err == nil {
		return
	}

	switch err {
	case flag.ErrHelp:
		os.Exit(2)
	default:
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func getEnvOrDefault(key string, defaultVal string) string {
	val, ok := os.LookupEnv(key)
	if ok {
		return val
	}
	return defaultVal
}

func run(args ...string) error {
	u, err := user.Current()
	if err != nil {
		panic(err)
	}

	app := keyman.NewCommand(
		"keyman [OPTIONS] ACTION ...",
		"keyman is the password management tool.",
	)
	app.Context(func(ctx keyman.Context) keyman.Context {
		return &appCtx{}
	})
	app.Flag(func(ctx keyman.Context, f *flag.FlagSet) {
		c := ctx.(*appCtx)
		p := getEnvOrDefault("KEYMAN_HOME", path.Join(u.HomeDir, ".keyman"))
		f.StringVar(&c.path, "p", p, "specify the keyman directory")
	})

	set := app.SubCommand(
		"set [-r REVISION] NAME",
		"set or revert password",
	)
	set.Context(func(ctx keyman.Context) keyman.Context {
		return &setCtx{
			appCtx: ctx.(*appCtx),
		}
	})
	set.Flag(func(ctx keyman.Context, f *flag.FlagSet) {
		c := ctx.(*setCtx)
		f.Uint64Var(&c.rev, "r", 0, "revert specified revision")
	})
	set.Handle(func(ctx keyman.Context, args ...string) error {
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "no name")
			return flag.ErrHelp
		}

		c := ctx.(*setCtx)
		k, err := keyman.Open(c.path)
		if err != nil {
			return err
		}
		defer k.Close()

		if c.rev > 0 {
			if err := k.Revert(args[0], c.rev); err != nil {
				return err
			}

			fmt.Printf("reverted: rev %d\n", c.rev)
			return nil
		}

		fmt.Fprint(os.Stderr, "Password: ")
		input, err := terminal.ReadPassword(int(syscall.Stdin))
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return err
		}

		rev, err := k.Set(args[0], input)
		if err != nil {
			return err
		}

		fmt.Printf("saved: rev %d\n", rev)
		return nil
	})

	info := app.SubCommand(
		"info NAME",
		"show password information",
	)
	info.Handle(func(ctx keyman.Context, args ...string) error {
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "no name")
			return flag.ErrHelp
		}

		c := ctx.(*appCtx)
		k, err := keyman.Open(c.path)
		if err != nil {
			return err
		}
		defer k.Close()

		rev, md, err := k.Metadata(args[0])
		if err != nil {
			return err
		}

		fmt.Printf("information %s\n", args[0])
		fmt.Println()
		fmt.Printf("revision: %d\n", rev)
		fmt.Printf("ctime: %d (%s)\n", md.CTime.Unix(), md.CTime.Local())
		fmt.Printf("length: %d\n", md.Length)
		fmt.Printf("hash (md5): %x\n", md.Md5)
		return nil
	})

	list := app.SubCommand(
		"list [-v]",
		"show password list",
	)
	list.Context(func(ctx keyman.Context) keyman.Context {
		return &listCtx{
			appCtx: ctx.(*appCtx),
		}
	})
	list.Flag(func(ctx keyman.Context, f *flag.FlagSet) {
		c := ctx.(*listCtx)
		f.BoolVar(&c.long, "l", false, "use a long listing format")
	})
	list.Handle(func(ctx keyman.Context, args ...string) error {
		c := ctx.(*listCtx)
		k, err := keyman.Open(c.path)
		if err != nil {
			return err
		}
		defer k.Close()

		var fn func(name string, rev uint64) error
		if c.long {
			fn = func(name string, rev uint64) error {
				_, err := fmt.Printf("%s (rev %d)\n", name, rev)
				return err
			}
		} else {
			fn = func(name string, rev uint64) error {
				_, err := fmt.Println(name)
				return err
			}

		}
		return k.ForEach(fn)
	})

	log := app.SubCommand(
		"log [OPTIONS] NAME",
		"show history",
	)
	log.Context(func(ctx keyman.Context) keyman.Context {
		return &logCtx{
			appCtx: ctx.(*appCtx),
		}
	})
	log.Flag(func(ctx keyman.Context, f *flag.FlagSet) {
		c := ctx.(*logCtx)
		f.Uint64Var(&c.limit, "n", 10, "limit the number of log to output")
	})
	log.Handle(func(ctx keyman.Context, args ...string) error {
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "no name")
			return flag.ErrHelp
		}

		c := ctx.(*logCtx)
		k, err := keyman.Open(c.path)
		if err != nil {
			return err
		}
		defer k.Close()

		cur, _, err := k.Metadata(args[0])
		if err != nil {
			return err
		}

		format := ""
		return k.History(args[0], c.limit, func(rev uint64, md *keyman.Metadata) error {
			if len(format) == 0 {
				n := fmt.Sprintf("%d", len(fmt.Sprintf("%d", rev)))
				format = "%" + n + "d: %x %d %s\n"
			}

			var err error
			if cur == rev {
				fmt.Printf("* ")
			} else {
				fmt.Printf("  ")
			}

			_, err = fmt.Printf(
				format,
				rev,
				md.Md5,
				md.Length,
				md.CTime.Local(),
			)
			return err
		})
	})

	delete := app.SubCommand(
		"delete NAME",
		"delete password (only delete reference)",
	)
	delete.Handle(func(ctx keyman.Context, args ...string) error {
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "no name")
			return flag.ErrHelp
		}

		c := ctx.(*appCtx)
		k, err := keyman.Open(c.path)
		if err != nil {
			return err
		}
		defer k.Close()

		return k.Delete(args[0])
	})

	in := app.SubCommand(
		"in NAME",
		"read password from stdin",
	)
	in.Handle(func(ctx keyman.Context, args ...string) error {
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "no name")
			return flag.ErrHelp
		}

		c := ctx.(*appCtx)
		k, err := keyman.Open(c.path)
		if err != nil {
			return err
		}
		defer k.Close()

		data, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return err
		}

		rev, err := k.Set(args[0], data)
		if err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "saved %d bytes: rev %d\n", len(data), rev)
		return nil
	})

	out := app.SubCommand(
		"out NAME",
		"write password to stdout",
	)
	out.Handle(func(ctx keyman.Context, args ...string) error {
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "no name")
			return flag.ErrHelp
		}

		c := ctx.(*appCtx)
		k, err := keyman.Open(c.path)
		if err != nil {
			return err
		}
		defer k.Close()

		data, err := k.Get(args[0])
		if err != nil {
			return err
		}

		os.Stdout.Write(data)
		return nil
	})

	tee := app.SubCommand(
		"tee NAME",
		"read password from stdin and write it to stdout",
	)
	tee.Handle(func(ctx keyman.Context, args ...string) error {
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "no name")
			return flag.ErrHelp
		}

		c := ctx.(*appCtx)
		k, err := keyman.Open(c.path)
		if err != nil {
			return err
		}
		defer k.Close()

		data, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return err
		}

		rev, err := k.Set(args[0], data)
		if err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "saved %d bytes: rev %d\n", len(data), rev)
		os.Stdout.Write(data)
		return nil
	})

	return app.Run(nil, args[1:]...)
}
