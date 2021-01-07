package app

import (
	"flag"
	"fmt"
)

// Usage returns the Usage function.
func Usage(cmd Command, f *flag.FlagSet) func() {
	return func() {
		w := f.Output()

		synopsis := cmd.Synopsis()
		fmt.Fprintf(w, "Usage: %s\n", synopsis[0])
		for _, s := range synopsis[1:] {
			fmt.Fprintf(w, "       %s\n", s)
		}

		fmt.Fprintln(w)
		fmt.Fprintln(w, "Options:")
		fmt.Fprintln(w, "  -h    show this message")
		f.PrintDefaults()
	}
}
