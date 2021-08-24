package shell

import (
	"fmt"
)

const (
	//ErrEmptyCommand is the error message when the provided command is empty
	ErrEmptyCommand string = "command is empty"
)

// executionFailedError is used to help formatting errors reported to the user
type executionFailedError struct {
	ErrCode int
	Stdout  string
	Stderr  string
	Command string
}

func (e *executionFailedError) Error() string {
	return errorMessage(e.ErrCode, e.Command, e.Stderr, e.Stdout)
}

func errorMessage(exitCode int, command, stdout, stderr string) string {
	return fmt.Sprintf("The shell 🐚 command %q failed with an exit code of %v and the following messages: \nstderr=\n%v\nstdout=\n%v\n", command, exitCode, stderr, stdout)
}
