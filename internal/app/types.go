package app

import "flag"

// Command is the interface of the CLI application.
type Command interface {
	Name() string
	Synopsis() []string
	Setup(*flag.FlagSet)
	Run(args []string) error
}
