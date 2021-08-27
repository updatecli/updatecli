package shell

import (
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/scm"
)

// Condition tests if the provided command (concatenated with the source) is executed with success
func (s *Shell) Condition(source string) (bool, error) {
	return s.condition(source, "")
}

// ConditionFromSCM tests if the provided command (concatenated with the source) is executed with success from the SCM root directory
func (s *Shell) ConditionFromSCM(source string, scm scm.Scm) (bool, error) {
	return s.condition(source, scm.GetDirectory())
}

func (s *Shell) condition(source, workingDir string) (bool, error) {
	customCommand := s.customCommand(source)

	cmdResult, err := s.executor.ExecuteCommand(command{
		Cmd: s.customCommand(source),
		Dir: workingDir,
	})

	if err != nil {
		return false, err
	}

	if cmdResult.ExitCode != 0 {
		logrus.Infof(errorMessage(cmdResult.ExitCode, customCommand, cmdResult.Stderr, cmdResult.Stdout))
		return false, nil
	}

	logrus.Infof("%v The shell 🐚 command %q successfully validated the condition.", result.SUCCESS, customCommand)

	return true, nil
}
