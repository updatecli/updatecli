package shell

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Condition tests if the provided command (concatenated with the source) is executed with success
func (s *Shell) Condition(source string, scm scm.ScmHandler, resultCondition *result.Condition) error {
	var workingDir string
	if scm != nil {
		workingDir = scm.GetDirectory()
	}

	// Ensure environment variable(s) are up to date
	// either it already has a value specified, or it retrieves
	// the value from the Updatecli process
	err := s.spec.Environments.Load()
	if err != nil {
		return err
	}

	// Provides the "UPDATECLI_PIPELINE_STAGE" environment variable set to "condition"
	// It's only purpose is to have at least one environment variable
	// so we don't fallback to use the current process environment as explained
	// on https://pkg.go.dev/os/exec#Cmd
	env := append(s.spec.Environments, Environment{
		Name:  CurrentStageVariableName,
		Value: "condition",
	})

	err = s.success.PreCommand(s.getWorkingDirPath(workingDir))
	if err != nil {
		return err
	}

	scriptFilename, err := newShellScript(s.appendSource(source))
	if err != nil {
		return fmt.Errorf("failed initializing source script - %s", err)
	}

	err = s.executeCommand(command{
		Cmd: s.interpreter + " " + scriptFilename,
		Dir: s.getWorkingDirPath(workingDir),
		Env: env.ToStringSlice(),
	})
	if err != nil {
		return fmt.Errorf("failed while running condition script - %s", err)
	}

	err = s.success.PostCommand(s.getWorkingDirPath(workingDir))
	if err != nil {
		return err
	}

	ok, err := s.success.ConditionResult()
	if err != nil {
		return err
	}

	switch ok {
	case true:
		resultCondition.Pass = true
		resultCondition.Result = result.SUCCESS
		resultCondition.Description = fmt.Sprintf("shell condition of type %q, passing", s.spec.ChangedIf.Kind)
	case false:
		resultCondition.Pass = false
		resultCondition.Result = result.SUCCESS
		resultCondition.Description = fmt.Sprintf("shell condition of type %q not passing", s.spec.ChangedIf)
	}

	return nil
}
