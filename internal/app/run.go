package app

import "os"

// Run executes the Command and returns the exit code.
func Run(cmd Command, args ...string) int {
	exec := &Executor{Stderr: os.Stderr}
	err := exec.Run(cmd, args...)
	return exec.HandleError(err)
}
