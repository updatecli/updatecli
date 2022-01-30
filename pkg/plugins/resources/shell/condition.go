package shell

import (
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

// Condition tests if the provided command (concatenated with the source) is executed with success
func (s *Shell) Condition(source string) (bool, error) {
	return s.condition(source, "")
}

// ConditionFromSCM tests if the provided command (concatenated with the source) is executed with success from the SCM root directory
func (s *Shell) ConditionFromSCM(source string, scm scm.ScmHandler) (bool, error) {
	return s.condition(source, scm.GetDirectory())
}

func (s *Shell) condition(source, workingDir string) (bool, error) {
	s.executeCommand(command{
		Cmd: s.appendSource(source),
		Dir: workingDir,
	})

	if s.result.ExitCode != 0 {
		return false, nil
	}

	return true, nil
}
