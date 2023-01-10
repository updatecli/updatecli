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
	// Ensure environment variable(s) are up to date
	// either it already has a value specified, or it retrieves
	// the value from the Updatecli process
	err := s.spec.Environments.Load()
	if err != nil {
		return false, err
	}

	// Provides the "UPDATECLI_PIPELINE_STAGE" environment variable set to "condition"
	// It's only purpose is to have at least one environment variable
	// so we don't fallback to use the current process environment as explained
	// on https://pkg.go.dev/os/exec#Cmd
	env := append(s.spec.Environments, Environment{
		Name:  CurrentStageVariableName,
		Value: "condition",
	})

	err = s.success.PreCommand()
	if err != nil {
		return false, err
	}

	s.executeCommand(command{
		Cmd: s.appendSource(source),
		Dir: workingDir,
		Env: env.ToStringSlice(),
	})

	err = s.success.PostCommand()
	if err != nil {
		return false, err
	}

	return s.success.ConditionResult()
}
