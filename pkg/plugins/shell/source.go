package shell

import (
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source returns the stdout of the shell command if its exit code is 0
// otherwise an error is returned with the content of stderr
func (s *Shell) Source(workingDir string) (string, error) {
	cmdResult, err := s.executor.ExecuteCommand(command{
		Cmd: s.spec.Command,
		Dir: workingDir,
	})

	if err != nil {
		return "", err
	}

	if cmdResult.ExitCode != 0 {
		return "", &executionFailedError{
			Command: s.spec.Command,
			ErrCode: cmdResult.ExitCode,
			Stdout:  cmdResult.Stdout,
			Stderr:  cmdResult.Stderr}
	}

	logrus.Infof("%v The shell 🐚 command %q ran successfully and retrieved the following source value: %q", result.SUCCESS, s.spec.Command, cmdResult.Stdout)

	return cmdResult.Stdout, nil
}
