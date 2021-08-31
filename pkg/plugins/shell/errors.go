package shell

import (
	"fmt"
)

type ErrEmptyCommand struct{}

func (e *ErrEmptyCommand) Error() string {
	return "Invalid spec for shell resource: command is empty."
}

// executionFailedError is used to help formatting errors reported to the user
type executionFailedError struct {
	ErrCode int
	Stdout  string
	Stderr  string
	Command string
}

func (e *executionFailedError) Error() string {
	return errorMessage(e.ErrCode, e.Command, e.Stdout, e.Stderr)
}

func errorMessage(exitCode int, command, stdout, stderr string) string {
	return fmt.Sprintf("The shell 🐚 command %q failed with an exit code of %v and the following messages: \nstderr=\n%v\nstdout=\n%v\n", command, exitCode, stderr, stdout)
}
