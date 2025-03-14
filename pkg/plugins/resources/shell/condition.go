package shell

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

// Condition tests if the provided command (concatenated with the source) is executed with success
func (s *Shell) Condition(source string, scm scm.ScmHandler) (pass bool, message string, err error) {
	var workingDir string
	if scm != nil {
		workingDir = scm.GetDirectory()
	}

	// Ensure environment variable(s) are up to date
	// either it already has a value specified, or it retrieves
	// the value from the Updatecli process
	ignoreEnvironmentsNotFound := s.spec.Environments == nil
	err = s.environments.Load(ignoreEnvironmentsNotFound)
	if err != nil {
		return false, "", err
	}

	// Provides the "UPDATECLI_PIPELINE_STAGE" environment variable set to "condition"
	// It's only purpose is to have at least one environment variable
	// so we don't fallback to use the current process environment as explained
	// on https://pkg.go.dev/os/exec#Cmd
	env := append(s.environments, Environment{
		Name:  CurrentStageVariableName,
		Value: "condition",
	})

	err = s.success.PreCommand(s.getWorkingDirPath(workingDir))
	if err != nil {
		return false, "", err
	}

	scriptFilename, err := newShellScript(s.appendSource(source))
	if err != nil {
		return false, "", fmt.Errorf("failed initializing source script - %s", err)
	}

	err = s.executeCommand(command{
		Cmd: s.interpreter + " " + scriptFilename,
		Dir: s.getWorkingDirPath(workingDir),
		Env: env.ToStringSlice(),
	})
	if err != nil {
		return false, "", fmt.Errorf("failed while running condition script - %s", err)
	}

	err = s.success.PostCommand(s.getWorkingDirPath(workingDir))
	if err != nil {
		return false, "", err
	}

	ok, err := s.success.ConditionResult()
	if err != nil {
		return false, "", err
	}

	if ok {
		return true, fmt.Sprintf("shell condition of type %q, passing", s.spec.ChangedIf.Kind), nil
	}

	return false, fmt.Sprintf("shell condition of type %q not passing", s.spec.ChangedIf.Kind), nil
}
