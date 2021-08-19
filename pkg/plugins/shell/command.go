package shell

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

type command struct {
	Cmd string
	Dir string
	Env []string
}

type commandResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

type commandExecutor interface {
	ExecuteCommand(cmd command) (commandResult, error)
}

type nativeCommandExecutor struct {
}

func (nce *nativeCommandExecutor) ExecuteCommand(inputCmd command) (commandResult, error) {
	if inputCmd.Cmd == "" {
		return commandResult{}, fmt.Errorf(ErrEmptyCommand)
	}

	var stdout, stderr bytes.Buffer
	cmdFields := strings.Fields(inputCmd.Cmd)
	command := exec.Command(cmdFields[0], cmdFields[1:]...) //nolint: gosec
	if inputCmd.Dir != "" {
		command.Dir = inputCmd.Dir
	}
	command.Stdout = &stdout
	command.Stderr = &stderr
	// Pass current environment to process and append the customized environment variables used internally by updatecli (such as DRY_RUN)
	command.Env = append(command.Env, inputCmd.Env...)
	err := command.Run()

	// Remove line returns from stdout
	out := strings.TrimSuffix(stdout.String(), "\n")

	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return commandResult{
			ExitCode: ee.ExitCode(),
			Stdout:   out,
			Stderr:   stderr.String(),
		}, nil
	}
	if err != nil {
		return commandResult{}, err
	}

	return commandResult{
		ExitCode: 0,
		Stdout:   out,
		Stderr:   stderr.String(),
	}, nil

}
