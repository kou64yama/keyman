package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"keyman"
	"keyman/pb"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/golang/protobuf/ptypes"
	"golang.org/x/crypto/ssh/terminal"
	"google.golang.org/grpc"
)

func client(conn *grpc.ClientConn, args ...string) error {
	cmd := args[0]
	args = args[1:]
	switch cmd {
	case "ls":
		return list(conn, args...)
	case "get":
		return get(conn, args...)
	case "set":
		return set(conn, args...)
	case "read":
		return read(conn, args...)
	case "rm":
		return remove(conn, args...)
	case "log":
		return showLog(conn, args...)
	}

	return fmt.Errorf("command provided but not defined: %s", cmd)
}

func list(conn *grpc.ClientConn, args ...string) error {
	var longOutput bool
	req := &pb.ListRequest{}
	f := flag.NewFlagSet(os.Args[0]+" ls", flag.ContinueOnError)
	f.BoolVar(&req.All, "a", false, "show all")
	f.BoolVar(&longOutput, "l", false, "long output")
	f.Usage = func() {
		usage := `
		Usage: %s [-l] [-a]

		  Show password list.
		`
		usage = strings.ReplaceAll(usage, "\n\t\t", "\n")
		usage = strings.TrimSpace(usage)
		fmt.Fprintf(f.Output(), usage, f.Name())
		fmt.Fprintln(f.Output())
		fmt.Fprintln(f.Output())
		fmt.Fprintln(f.Output(), "Options:")
		f.PrintDefaults()
	}
	if err := f.Parse(args); err != nil {
		return errExitUsage
	}

	client := pb.NewKeymanClient(conn)
	stream, err := client.List(context.Background(), req)
	if err != nil {
		return err
	}

	for true {
		meta, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if longOutput {
			t, err := ptypes.Timestamp(meta.GetTime())
			if err != nil {
				return err
			}
			fmt.Printf(
				"%s %6d %s %s\n",
				meta.GetHash()[:7],
				meta.GetSize(),
				t.Local().Format(time.Stamp),
				meta.GetName(),
			)
		} else {
			fmt.Println(meta.GetName())
		}
	}
	return nil
}

func get(conn *grpc.ClientConn, args ...string) error {
	var noNewline bool
	req := &pb.GetRequest{}
	f := flag.NewFlagSet(os.Args[0]+" get", flag.ContinueOnError)
	f.BoolVar(&noNewline, "n", false, "do not print the trailing newline")
	f.Usage = func() {
		usage := `
		Usage: %s [-n] <NAME> [REV]

		  Print password.
		`
		usage = strings.ReplaceAll(usage, "\n\t\t", "\n")
		usage = strings.TrimSpace(usage)
		fmt.Fprintf(f.Output(), usage, f.Name())
		fmt.Fprintln(f.Output())
		fmt.Fprintln(f.Output())
		fmt.Fprintln(f.Output(), "Options:")
		f.PrintDefaults()
	}
	if err := f.Parse(args); err != nil {
		return errExitUsage
	}
	req.Name = f.Arg(0)
	if f.NArg() == 2 {
		req.Hash = f.Arg(1)
	}

	client := pb.NewKeymanClient(conn)
	stream, err := client.Get(context.Background(), req)
	if err != nil {
		return err
	}

	for true {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		os.Stdout.Write(res.GetChunk())
	}
	if !noNewline {
		os.Stdout.Write([]byte{'\n'})
	}
	return nil
}

func set(conn *grpc.ClientConn, args ...string) error {
	f := flag.NewFlagSet(os.Args[0]+" set", flag.ContinueOnError)
	f.Usage = func() {
		usage := `
		Usage: %s <NAME>

		  Input password.
		`
		usage = strings.ReplaceAll(usage, "\n\t\t", "\n")
		usage = strings.TrimSpace(usage)
		fmt.Fprintf(f.Output(), usage, f.Name())
		fmt.Fprintln(f.Output())
		fmt.Fprintln(f.Output())
		fmt.Fprintln(f.Output(), "Options:")
		f.PrintDefaults()
	}
	if err := f.Parse(args); err != nil {
		return errExitUsage
	}

	fmt.Fprint(os.Stderr, "Enter password: ")
	b, err := terminal.ReadPassword(syscall.Stdin)
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return err
	}

	client := pb.NewKeymanClient(conn)
	stream, err := client.Set(context.Background())
	if err != nil {
		return err
	}

	if err := stream.Send(&pb.SetRequest{Value: &pb.SetRequest_Name{Name: f.Arg(0)}}); err != nil {
		return err
	}
	r := bytes.NewReader(b)
	buf := make([]byte, keyman.ChunkSize)
	for true {
		size, err := r.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if err := stream.Send(&pb.SetRequest{Value: &pb.SetRequest_Chunk{Chunk: buf[:size]}}); err != nil {
			return err
		}
	}

	meta, err := stream.CloseAndRecv()
	if err != nil {
		return err
	}
	fmt.Println(meta.Hash[:7])
	return nil
}

func read(conn *grpc.ClientConn, args ...string) error {
	f := flag.NewFlagSet(os.Args[0]+" read", flag.ContinueOnError)
	f.Usage = func() {
		usage := `
		Usage: %s <NAME>

		  Read password from STDIN.
		`
		usage = strings.ReplaceAll(usage, "\n\t\t", "\n")
		usage = strings.TrimSpace(usage)
		fmt.Fprintf(f.Output(), usage, f.Name())
		fmt.Fprintln(f.Output())
		fmt.Fprintln(f.Output())
		fmt.Fprintln(f.Output(), "Options:")
		f.PrintDefaults()
	}
	if err := f.Parse(args); err != nil {
		return errExitUsage
	}

	client := pb.NewKeymanClient(conn)
	stream, err := client.Set(context.Background())
	if err != nil {
		return err
	}

	if err := stream.Send(&pb.SetRequest{Value: &pb.SetRequest_Name{Name: f.Arg(0)}}); err != nil {
		return err
	}
	buf := make([]byte, keyman.ChunkSize)
	for true {
		size, err := os.Stdin.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if err := stream.Send(&pb.SetRequest{Value: &pb.SetRequest_Chunk{Chunk: buf[:size]}}); err != nil {
			return err
		}
	}

	meta, err := stream.CloseAndRecv()
	if err != nil {
		return err
	}
	fmt.Println(meta.Hash[:7])
	return nil
}

func remove(conn *grpc.ClientConn, args ...string) error {
	f := flag.NewFlagSet(os.Args[0]+" rm", flag.ContinueOnError)
	f.Usage = func() {
		usage := `
		Usage: %s <NAME>

		  Remove password.
		`
		usage = strings.ReplaceAll(usage, "\n\t\t", "\n")
		usage = strings.TrimSpace(usage)
		fmt.Fprintf(f.Output(), usage, f.Name())
		fmt.Fprintln(f.Output())
		fmt.Fprintln(f.Output())
		fmt.Fprintln(f.Output(), "Options:")
		f.PrintDefaults()
	}
	if err := f.Parse(args); err != nil {
		return errExitUsage
	}

	client := pb.NewKeymanClient(conn)
	stream, err := client.Set(context.Background())
	if err != nil {
		return err
	}

	if err := stream.Send(&pb.SetRequest{Value: &pb.SetRequest_Name{Name: f.Arg(0)}}); err != nil {
		return err
	}

	meta, err := stream.CloseAndRecv()
	if err != nil {
		return err
	}
	fmt.Println(meta.Hash[:7])
	return nil
}

func showLog(conn *grpc.ClientConn, args ...string) error {
	req := &pb.LogRequest{}
	f := flag.NewFlagSet(os.Args[0]+" log", flag.ContinueOnError)
	f.Usage = func() {
		usage := `
		Usage: %s [-n LIMIT] <NAME>

		  Show password log.
		`
		usage = strings.ReplaceAll(usage, "\n\t\t", "\n")
		usage = strings.TrimSpace(usage)
		fmt.Fprintf(f.Output(), usage, f.Name())
		fmt.Fprintln(f.Output())
		fmt.Fprintln(f.Output())
		fmt.Fprintln(f.Output(), "Options:")
		f.PrintDefaults()
	}
	f.Uint64Var(&req.Limit, "n", 0, "limit the number of log entries")
	if err := f.Parse(args); err != nil {
		return errExitUsage
	}
	req.Name = f.Arg(0)

	client := pb.NewKeymanClient(conn)
	ctx := context.Background()
	stream, err := client.Log(ctx, req)
	if err != nil {
		return err
	}
	for true {
		meta, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		t, err := ptypes.Timestamp(meta.GetTime())
		if err != nil {
			return err
		}
		fmt.Printf(
			"%s %s %6d\n",
			meta.GetHash()[:7],
			t.Local().Format(time.RFC822),
			meta.GetSize(),
		)
	}
	return nil
}
