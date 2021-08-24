package shell

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/scm"
)

func (s *Shell) Target(source string, dryRun bool) (bool, error) {
	changed, _, _, err := s.target(source, "", dryRun)
	return changed, err
}

func (s *Shell) TargetFromSCM(source string, scm scm.Scm, dryRun bool) (changed bool, commands []string, message string, err error) {
	return s.target(source, scm.GetDirectory(), dryRun)
}

// Target executes the provided command (concatenated with the source) to apply the change.
//	The command is expected, if it changes something, to print the new value to the stdout
//	- An exit code of 0 and an empty stdout means: "successful command and no change"
//	- An exit code of 0 and something on the stdout means: "successful command with a changed value"
//	- Any other exit code means "failed command with no change"
// The environment variable 'DRY_RUN' is set to true or false based on the input parameter (e.g. 'updatecli diff' or 'apply'?)
func (s *Shell) target(source, workingDir string, dryRun bool) (changed bool, commands []string, message string, err error) {
	customCommand := s.customCommand(source)

	cmdResult, err := s.executor.ExecuteCommand(command{
		Cmd: customCommand,
		Dir: workingDir,
		Env: []string{fmt.Sprintf("DRY_RUN=%v", dryRun)},
	})

	if err != nil {
		return false, commands, message, err
	}

	if cmdResult.ExitCode != 0 {
		return false, commands, message, &executionFailedError{
			Command: customCommand,
			ErrCode: cmdResult.ExitCode,
			Stdout:  cmdResult.Stdout,
			Stderr:  cmdResult.Stderr,
		}
	}

	commands = append(commands, customCommand)

	if cmdResult.Stdout == "" {
		logrus.Infof("%v The shell 🐚 command %q ran successfully with no change.", result.SUCCESS, customCommand)
		return false, commands, message, nil
	}

	message = fmt.Sprintf("%v The shell 🐚 command %q ran successfully and reported the following change: %q.", result.CHANGED, customCommand, cmdResult.Stdout)
	logrus.Infof(message)

	return true, commands, message, nil
}
