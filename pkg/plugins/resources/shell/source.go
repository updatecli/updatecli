package shell

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source returns the stdout of the shell command if its exit code is 0
// otherwise an error is returned with the content of stderr
func (s *Shell) Source(workingDir string, resultSource *result.Source) error {

	// Ensure environment variable(s) are up to date
	// either it already has a value specified, or it retrieves
	// the value from the Updatecli process
	ignoreEnvironmentsNotFound := s.spec.Environments == nil
	err := s.environments.Load(ignoreEnvironmentsNotFound)
	if err != nil {
		return &ExecutionFailedError{}
	}

	// Provides the "UPDATECLI_PIPELINE_STAGE" environment variable set to "source"
	// It's only purpose is to have at least one environment variable
	// so we don't fallback to use the current process environment as explained
	// on https://pkg.go.dev/os/exec#Cmd
	// Provides the "DRY_RUN" environment variable to the shell command (true if "diff", false if "apply")
	sourceStageValue := "source"
	env := append(s.environments, Environment{
		Name:  CurrentStageVariableName,
		Value: &sourceStageValue,
	})

	// PreCommand is executed to collect information before running the shell command
	// so we could collect information needed to validate that a command successfully as expected
	err = s.success.PreCommand(s.getWorkingDirPath(workingDir))
	if err != nil {
		return fmt.Errorf("running precommand: %w", err)
	}

	scriptFilename, err := newShellScript(s.spec.Command)
	if err != nil {
		return fmt.Errorf("initializing source script: %w", err)
	}

	err = s.executeCommand(command{
		Cmd: s.interpreter + " " + scriptFilename,
		Dir: s.getWorkingDirPath(workingDir),
		Env: env.ToStringSlice(),
	})
	if err != nil {
		return fmt.Errorf("running source script: %w", err)
	}

	// PostCommand is executed to collect information after running the shell command
	// so we could collect information needed to validate that a command successfully as expected
	err = s.success.PostCommand(s.getWorkingDirPath(workingDir))
	if err != nil {
		return fmt.Errorf("running postcommand: %w", err)
	}

	return s.success.SourceResult(resultSource)
}
