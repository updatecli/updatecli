package shell

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
)

type command struct {
	Shell string
	Cmd   string
	Dir   string
	Env   []string
}

type commandResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Cmd      string
}

type commandExecutor interface {
	ExecuteCommand(cmd command) (commandResult, error)
}

type nativeCommandExecutor struct{}

func (nce *nativeCommandExecutor) ExecuteCommand(inputCmd command) (commandResult, error) {
	var stdout, stderr bytes.Buffer

	logrus.Debugf("\tcommand: %s %s\n", inputCmd.Shell, inputCmd.Cmd)

	cmdFields := strings.Fields(inputCmd.Cmd)
	command := exec.Command(cmdFields[0], cmdFields[1:]...) //nolint: gosec

	command.Dir = inputCmd.Dir
	command.Stdout = &stdout
	command.Stderr = &stderr
	// Pass current environment to process and append the customized environment variables used internally by updatecli (such as DRY_RUN)
	command.Env = append(command.Env, inputCmd.Env...)
	err := command.Run()

	// Display environment variables in debug mode
	logrus.Debugf("Environment variables\n")
	for _, env := range command.Env {
		logrus.Debugf("\t* %s\n", env)
	}

	// Remove line returns from stdout
	out := strings.TrimSuffix(stdout.String(), "\n")

	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return commandResult{
			ExitCode: ee.ExitCode(),
			Stdout:   out,
			Stderr:   stderr.String(),
			Cmd:      inputCmd.Cmd,
		}, nil
	}
	if err != nil {
		return commandResult{}, err
	}

	return commandResult{
		ExitCode: 0,
		Stdout:   out,
		Stderr:   stderr.String(),
		Cmd:      inputCmd.Cmd,
	}, nil
}
