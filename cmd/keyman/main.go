package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"google.golang.org/grpc"
)

type config struct {
	dir    string
	socket string
	debug  bool
}

var (
	version      = "0.0.0"
	errExitUsage = errors.New("exit usage")
)

func main() {
	c := &config{}
	err := run(c, os.Args[1:]...)
	if err == errExitUsage {
		os.Exit(2)
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		os.Exit(exitErr.ExitCode())
	}
	if err != nil {
		panic(err)
	}
}

func dial(c *config) func(context.Context, string) (net.Conn, error) {
	return func(ctx context.Context, socket string) (net.Conn, error) {
		conn, err := net.Dial("unix", socket)
		if err != nil {
			daemon(c)
			conn, err = net.Dial("unix", socket)
		}
		return conn, err
	}
}

func run(c *config, args ...string) error {
	var daemonFlag bool
	f := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	f.BoolVar(&daemonFlag, "d", false, "daemon mode")
	f.StringVar(&c.dir, "p", "", "path to keyman root directory")
	f.StringVar(&c.socket, "s", "", "path to socket")
	f.BoolVar(&c.debug, "x", false, "debug mode")
	f.Usage = func() {
		usage := `
		Usage: %s [-h]
		       %s [-d] [-p DATA_PATH] [-s SOCKET_PATH] [-x]
		       %s <COMMAND> [ARGUMENTS...]

		  Keyman is the password manager.

		Commands:
		  ls     show password list
		  get    print password
		  set    input password
		  read   read password from STDIN
		  rm     remove password
		  log    show password log
		`
		usage = strings.ReplaceAll(usage, "\n\t\t", "\n")
		usage = strings.TrimSpace(usage)
		fmt.Fprintf(f.Output(), usage, f.Name(), f.Name(), f.Name())
		fmt.Fprintln(f.Output())
		fmt.Fprintln(f.Output())
		fmt.Fprintln(f.Output(), "Options:")
		f.PrintDefaults()
	}
	if err := f.Parse(args); err != nil {
		return errExitUsage
	}

	home := os.Getenv("KEYMAN_HOME")
	if len(home) == 0 {
		dir, err := os.UserConfigDir()
		if err != nil {
			return err
		}
		home = filepath.Join(dir, "keyman")
	}

	if len(c.dir) == 0 {
		c.dir = filepath.Join(home, "data")
	}

	if len(c.socket) == 0 {
		c.socket = os.Getenv("KEYMAN_SOCK")
	}
	if len(c.socket) == 0 {
		c.socket = filepath.Join(home, "keyman.sock")
	}

	if f.NArg() == 0 {
		if !daemonFlag {
			return server(c)
		}
		cmd, err := daemon(c)
		if err != nil {
			return err
		}
		fmt.Printf("KEYMAN_SOCK=%s; export KEYMAN_SOCK;\n", c.socket)
		fmt.Printf("KEYMAN_PID=%d; export KEYMAN_PID;\n", cmd.Process.Pid)
		fmt.Printf("echo Keyman pid %d;\n", cmd.Process.Pid)
		return nil
	}

	conn, err := grpc.Dial(c.socket, grpc.WithInsecure(), grpc.WithContextDialer(dial(c)))
	if err != nil {
		return err
	}
	defer conn.Close()
	return client(conn, f.Args()...)
}
