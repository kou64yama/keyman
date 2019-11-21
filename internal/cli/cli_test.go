package cli_test

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/kou64yama/keyman/internal/cli"
)

func TestNewContext(t *testing.T) {
	t.Helper()

	c := cli.NewContext()
	if c.Stdin != os.Stdin {
		t.Errorf("got %v, want %v", c.Stdin, os.Stdin)
	}
	if c.Stdout != os.Stdout {
		t.Errorf("got %v, want %v", c.Stdout, os.Stdout)
	}
	if c.Stderr != os.Stderr {
		t.Errorf("got %v, want %v", c.Stderr, os.Stderr)
	}
}

func TestExitCode(t *testing.T) {
	tests := []struct {
		err  error
		code int
	}{
		{err: nil, code: 0},
		{err: flag.ErrHelp, code: 2},
		{err: errors.New("others"), code: 1},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("returns %d for %v", tt.code, tt.err)
		t.Run(name, func(t *testing.T) {
			t.Helper()

			code := cli.ExitCode(tt.err)
			if code != tt.code {
				t.Errorf("got %d, want %d", code, tt.code)
			}
		})
	}
}
