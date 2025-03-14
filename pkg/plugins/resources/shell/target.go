package shell

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (s *Shell) Target(source string, scm scm.ScmHandler, dryRun bool, resultTarget *result.Target) error {
	getDir := ""
	if scm != nil {
		getDir = scm.GetDirectory()
	}

	err := s.target(source, getDir, dryRun, resultTarget)
	if err != nil {
		return err
	}

	if scm != nil {
		// Once the changes have been applied inside the scm's temp directory, then we have to get the list of these changes
		resultTarget.Files, err = scm.GetChangedFiles(scm.GetDirectory())
		if err != nil {
			return err
		}
	}

	return nil
}

// Target executes the provided command (concatenated with the source) to apply the change.
// The command is expected, if it changes something, to print the new value to the stdout
//   - An exit code of 0 and an empty stdout means: "successful command and no change"
//   - An exit code of 0 and something on the stdout means: "successful command with a changed value"
//   - Any other exit code means "failed command with no change"
//
// The environment variable 'DRY_RUN' is set to true or false based on the input parameter (e.g. 'updatecli diff' or 'apply'?)
func (s *Shell) target(source, workingDir string, dryRun bool, resultTarget *result.Target) error {

	// Ensure environment variable(s) are up to date
	// either it already has a value specified, or it retrieves
	// the value from the Updatecli process
	ignoreEnvironmentsNotFound := s.spec.Environments == nil
	err := s.environments.Load(ignoreEnvironmentsNotFound)
	if err != nil {
		return err
	}

	// Provides the "UPDATECLI_PIPELINE_STAGE" environment variable set to "target"
	// It's only purpose is to have at least one environment variable
	// so we don't fallback to use the current process environment as explained
	// on https://pkg.go.dev/os/exec#Cmd
	env := append(s.environments, Environment{
		Name:  CurrentStageVariableName,
		Value: "target",
	})

	// Provides the "DRY_RUN" environment variable to the shell command (true if "diff", false if "apply")
	env = append(env, Environment{
		Name:  DryRunVariableName,
		Value: fmt.Sprintf("%v", dryRun),
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
		return fmt.Errorf("failed while running target script - %s", err)
	}

	err = s.success.PostCommand(s.getWorkingDirPath(workingDir))
	if err != nil {
		return err
	}

	resultTarget.Changed, err = s.success.TargetResult()

	if err != nil {
		return &ExecutionFailedError{}
	}

	resultTarget.Description = fmt.Sprintf("ran shell command %q", s.appendSource(source))

	if !resultTarget.Changed {
		resultTarget.NewInformation = source
		resultTarget.Information = source

		resultTarget.Result = result.SUCCESS
		logrus.Info("No change detected")
		return nil
	}

	resultTarget.Result = result.ATTENTION
	resultTarget.Information = "unknown"

	return nil
}
