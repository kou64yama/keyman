package keyman

import (
	"bytes"
	"flag"
	"strings"
	"testing"
)

type stdStub struct {
	in  *bytes.Buffer
	out *bytes.Buffer
	err *bytes.Buffer
}

func mockCommand(synopsis string, help string) (Command, *stdStub) {
	s := &stdStub{
		in:  bytes.NewBuffer([]byte{}),
		out: bytes.NewBuffer([]byte{}),
		err: bytes.NewBuffer([]byte{}),
	}
	c := &cmd{
		name:     strings.Split(synopsis, " ")[0],
		synopsis: synopsis,
		help:     help,
		subMap:   make(map[string]Command),
		stdin:    s.in,
		stdout:   s.out,
		stderr:   s.err,
	}
	return c, s
}

func TestCommandRunHelp(t *testing.T) {
	cmd, std := mockCommand("cmd", "description.")
	cmd.Flag(func(ctx Context, f *flag.FlagSet) {
		f.String("foo", "foo", "help foo")
		f.String("bar", "bar", "help bar")
	})
	cmd.SubCommand("baz", "help baz")
	cmd.SubCommand("qux", "help qux")
	cmd.Run(nil, "-h")

	t.Log(std.err.String())
	if std.err.Len() == 0 {
		t.Fatal()
	}
}

func TestCommandRun(t *testing.T) {
	cmd := NewCommand("cmd", "")

	var called []string
	cmd.Handle(func(ctx Context, args ...string) error {
		called = args
		return nil
	})

	cmd.Run(nil, "foo", "bar")
	if called[0] != "foo" || called[1] != "bar" {
		t.Fatal("arguments:", called)
	}
}

func TestCommandSubCommandRun(t *testing.T) {
	cmd := NewCommand("cmd", "")
	sub := cmd.SubCommand("sub", "")

	var called []string
	sub.Handle(func(ctx Context, args ...string) error {
		called = args
		return nil
	})

	cmd.Run(nil, "sub", "foo", "bar")
	if called[0] != "foo" || called[1] != "bar" {
		t.Fatal("arguments:", called)
	}
}

func TestCommandNoArguments(t *testing.T) {
	cmd, _ := mockCommand("cmd", "")
	err := cmd.Run(nil)
	if err != flag.ErrHelp {
		t.Fail()
	}
}

func TestCommandInvalidArguments(t *testing.T) {
	cmd, _ := mockCommand("cmd", "")
	err := cmd.Run(nil, "sub")
	if err != flag.ErrHelp {
		t.Fail()
	}
}

type mockCtx struct {
	Command
}

func TestCommandContext(t *testing.T) {
	cmd := NewCommand("cmd", "")
	cmd.Context(func(ctx Context) Context { return &mockCtx{} })

	var called Context
	cmd.Handle(func(ctx Context, args ...string) error {
		called = ctx
		return nil
	})

	cmd.Run(nil)
	_, ok := called.(*mockCtx)
	if !ok {
		t.Fatal()
	}
}
